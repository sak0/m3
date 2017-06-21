// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package msgpack

import (
	"errors"
	"sync"

	"github.com/m3db/m3cluster/services"
	"github.com/m3db/m3metrics/metric/unaggregated"
	"github.com/m3db/m3metrics/policy"
	"github.com/m3db/m3metrics/protocol/msgpack"
	"github.com/m3db/m3x/errors"
	"github.com/m3db/m3x/log"
)

var (
	errInstanceWriterClosed   = errors.New("instance writer is closed")
	errUnrecognizedMetricType = errors.New("unrecognized metric type")
)

type instanceWriter interface {
	// Write writes a metric alongside its policies list for a given shard.
	Write(shard uint32, mu unaggregated.MetricUnion, pl policy.PoliciesList) error

	// Flush flushes any buffered metrics.
	Flush() error

	// Close closes the writer.
	Close() error
}

type newLockedEncoderFn func(msgpack.BufferedEncoderPool) *lockedEncoder

type writer struct {
	sync.RWMutex

	log               xlog.Logger
	flushSize         int
	maxTimerBatchSize int
	encoderPool       msgpack.BufferedEncoderPool
	queue             instanceQueue

	closed             bool
	encodersByShard    map[uint32]*lockedEncoder
	newLockedEncoderFn newLockedEncoderFn
}

func newInstanceWriter(instance services.PlacementInstance, opts ServerOptions) instanceWriter {
	w := &writer{
		log:               opts.InstrumentOptions().Logger(),
		flushSize:         opts.FlushSize(),
		maxTimerBatchSize: opts.MaxTimerBatchSize(),
		encoderPool:       opts.BufferedEncoderPool(),
		queue:             newInstanceQueue(instance, opts),
		encodersByShard:   make(map[uint32]*lockedEncoder),
	}
	w.newLockedEncoderFn = newLockedEncoder
	return w
}

func (w *writer) Write(
	shard uint32,
	mu unaggregated.MetricUnion,
	pl policy.PoliciesList,
) error {
	w.RLock()
	if w.closed {
		w.RUnlock()
		return errInstanceWriterClosed
	}
	encoder, exists := w.encodersByShard[shard]
	if exists {
		err := w.encodeWithLock(encoder, mu, pl)
		w.RUnlock()
		return err
	}
	w.RUnlock()

	w.Lock()
	if w.closed {
		w.Unlock()
		return errInstanceWriterClosed
	}
	encoder, exists = w.encodersByShard[shard]
	if exists {
		err := w.encodeWithLock(encoder, mu, pl)
		w.Unlock()
		return err
	}
	encoder = w.newLockedEncoderFn(w.encoderPool)
	w.encodersByShard[shard] = encoder
	err := w.encodeWithLock(encoder, mu, pl)
	w.Unlock()

	return err
}

func (w *writer) Flush() error {
	w.RLock()
	if w.closed {
		w.RUnlock()
		return errInstanceWriterClosed
	}
	err := w.flushWithLock()
	w.RUnlock()
	return err
}

func (w *writer) Close() error {
	w.Lock()
	defer w.Unlock()

	if w.closed {
		return errInstanceWriterClosed
	}
	w.closed = true
	w.flushWithLock()
	return w.queue.Close()
}

func (w *writer) flushWithLock() error {
	multiErr := xerrors.NewMultiError()
	for _, encoder := range w.encodersByShard {
		encoder.Lock()
		bufferedEncoder := encoder.Encoder()
		buffer := bufferedEncoder.Buffer()
		if buffer.Len() == 0 {
			encoder.Unlock()
			continue
		}
		newBufferedEncoder := w.encoderPool.Get()
		newBufferedEncoder.Reset()
		encoder.Reset(newBufferedEncoder)
		encoder.Unlock()
		if err := w.queue.Enqueue(bufferedEncoder); err != nil {
			multiErr = multiErr.Add(err)
		}
	}
	return multiErr.FinalError()
}

func (w *writer) encodeWithLock(
	encoder *lockedEncoder,
	mu unaggregated.MetricUnion,
	pl policy.PoliciesList,
) error {
	encoder.Lock()

	var (
		bufferedEncoder = encoder.Encoder()
		buffer          = bufferedEncoder.Buffer()
		sizeBefore      = buffer.Len()
		err             error
	)
	switch mu.Type {
	case unaggregated.CounterType:
		cp := unaggregated.CounterWithPoliciesList{
			Counter:      mu.Counter(),
			PoliciesList: pl,
		}
		err = encoder.EncodeCounterWithPoliciesList(cp)
	case unaggregated.BatchTimerType:
		// If there is no limit on the timer batch size, write the full batch.
		if w.maxTimerBatchSize == 0 {
			btp := unaggregated.BatchTimerWithPoliciesList{
				BatchTimer:   mu.BatchTimer(),
				PoliciesList: pl,
			}
			err = encoder.EncodeBatchTimerWithPoliciesList(btp)
			break
		}

		// Otherwise, honor maximum timer batch size.
		var (
			batchTimer     = mu.BatchTimer()
			timerValues    = batchTimer.Values
			numTimerValues = len(timerValues)
			start, end     int
		)
		for start = 0; start < numTimerValues && err == nil; start = end {
			end = start + w.maxTimerBatchSize
			if end > numTimerValues {
				end = numTimerValues
			}
			btp := unaggregated.BatchTimerWithPoliciesList{
				BatchTimer: unaggregated.BatchTimer{
					ID:     batchTimer.ID,
					Values: timerValues[start:end],
				},
				PoliciesList: pl,
			}
			err = encoder.EncodeBatchTimerWithPoliciesList(btp)
		}
	case unaggregated.GaugeType:
		gp := unaggregated.GaugeWithPoliciesList{
			Gauge:        mu.Gauge(),
			PoliciesList: pl,
		}
		err = encoder.EncodeGaugeWithPoliciesList(gp)
	default:
		err = errUnrecognizedMetricType
	}

	if err != nil {
		w.log.WithFields(
			xlog.NewLogField("metric", mu),
			xlog.NewLogField("policies", pl),
			xlog.NewLogErrField(err),
		).Error("encode metric with policies error")
		// Rewind buffer and clear out the encoder error.
		buffer.Truncate(sizeBefore)
		encoder.Reset(bufferedEncoder)
		encoder.Unlock()
		return err
	}

	// If the buffer size is not big enough, do nothing.
	sizeAfter := buffer.Len()
	if sizeAfter < w.flushSize {
		encoder.Unlock()
		return nil
	}

	// Otherwise we get a new buffer from pool, copy the bytes exceeding the
	// flush size to it, swap the new buffer with the old one, and flush out
	// the old buffer.
	newBufferedEncoder := w.encoderPool.Get()
	newBufferedEncoder.Reset()
	if sizeBefore > 0 {
		data := bufferedEncoder.Bytes()
		newBufferedEncoder.Buffer().Write(data[sizeBefore:sizeAfter])
		buffer.Truncate(sizeBefore)
	}
	encoder.Reset(newBufferedEncoder)
	encoder.Unlock()

	// Write out the buffered data.
	return w.queue.Enqueue(bufferedEncoder)
}

type lockedEncoder struct {
	sync.Mutex
	msgpack.UnaggregatedEncoder
}

func newLockedEncoder(p msgpack.BufferedEncoderPool) *lockedEncoder {
	bufferedEncoder := p.Get()
	bufferedEncoder.Reset()
	encoder := msgpack.NewUnaggregatedEncoder(bufferedEncoder)
	return &lockedEncoder{UnaggregatedEncoder: encoder}
}

type refCountedWriter struct {
	refCount
	instanceWriter
}

func newRefCountedWriter(instance services.PlacementInstance, opts ServerOptions) *refCountedWriter {
	rcWriter := &refCountedWriter{instanceWriter: newInstanceWriter(instance, opts)}
	rcWriter.refCount.SetDestructor(rcWriter.Close)
	return rcWriter
}

func (rcWriter *refCountedWriter) Close() {
	rcWriter.instanceWriter.Close()
}