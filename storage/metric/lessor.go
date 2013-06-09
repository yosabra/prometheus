// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric

import (
	dto "github.com/prometheus/prometheus/model/generated"
)

type labelNameLessor chan *dto.LabelName

func (l labelNameLessor) lease() *dto.LabelName {
	select {
	case v := <-l:
		return v
	default:
		return &dto.LabelName{}
	}
}

func (l labelNameLessor) credit(v *dto.LabelName) {
	v.Reset()

	select {
	case l <- v:
	default:
	}
}

func (l labelNameLessor) close() {
	close(l)

	for _ = range l {
	}
}

type fingerprintCollectionLessor chan *dto.FingerprintCollection

func (l fingerprintCollectionLessor) lease() *dto.FingerprintCollection {
	select {
	case v := <-l:
		return v
	default:
		return &dto.FingerprintCollection{}
	}
}

func (l fingerprintCollectionLessor) credit(v *dto.FingerprintCollection) {
	v.Reset()

	select {
	case l <- v:
	default:
	}
}

func (l fingerprintCollectionLessor) close() {
	close(l)

	for _ = range l {
	}
}

type labelPairLessor chan *dto.LabelPair

func (l labelPairLessor) lease() *dto.LabelPair {
	select {
	case v := <-l:
		return v
	default:
		return &dto.LabelPair{}
	}
}

func (l labelPairLessor) credit(v *dto.LabelPair) {
	v.Reset()

	select {
	case l <- v:
	default:
	}
}

func (l labelPairLessor) close() {
	close(l)

	for _ = range l {
	}
}

type fingerprintLessor chan *dto.Fingerprint

func (l fingerprintLessor) lease() *dto.Fingerprint {
	select {
	case v := <-l:
		return v
	default:
		return &dto.Fingerprint{}
	}
}

func (l fingerprintLessor) credit(v *dto.Fingerprint) {
	v.Reset()

	select {
	case l <- v:
	default:
	}
}

func (l fingerprintLessor) close() {
	close(l)

	for _ = range l {
	}
}

type highWatermarkLessor chan *dto.MetricHighWatermark

func (l highWatermarkLessor) lease() *dto.MetricHighWatermark {
	select {
	case v := <-l:
		return v
	default:
		return &dto.MetricHighWatermark{}
	}
}

func (l highWatermarkLessor) credit(v *dto.MetricHighWatermark) {
	v.Reset()

	select {
	case l <- v:
	default:
	}
}

func (l highWatermarkLessor) close() {
	close(l)

	for _ = range l {
	}
}
