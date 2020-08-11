// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import "testing"

func TestStats_Update(t *testing.T) {
	s := newStats()
	s.update(positive, age20To29, 1)
	s.update(negative, age20To29, 2)
	s.update(positive, age30To39, 3)
	s.update(otherResult, age40To49, 3)
	s.update(negative, age0To9, 4)
	s.update(positive, age60To69, 5)

	if s.pos != 3 {
		t.Errorf("pos = %v; want 3", s.pos)
	}
	if s.neg != 2 {
		t.Errorf("neg = %v; want 2", s.pos)
	}
	if s.other != 1 {
		t.Errorf("other = %v; want 1", s.pos)
	}

	if v := s.agePos[age20To29]; v != 1 {
		t.Errorf("agePos[age20To29] = %v; want 1", v)
	}
	if v := s.agePos[age40To49]; v != 0 {
		t.Errorf("agePos[age40To49] = %v; want 0", v)
	}

	if v := s.delayPct(50); v != 3 {
		t.Errorf("delayPct(50) = %v; want 3", v)
	}
	if v := s.posDelayPct(0); v != 1 {
		t.Errorf("posDelayPct(0) = %v; want 1", v)
	}
	if v := s.negDelayPct(0); v != 2 {
		t.Errorf("negDelayPct(50) = %v; want 2", v)
	}
}

func TestHist_Percentile(t *testing.T) {
	h := newHist(7)
	h.inc(0)
	h.inc(2)
	h.inc(7)
	if v := h.percentile(0); v != 0 {
		t.Errorf("percentile(0) = %v; want 0", v)
	}
	if v := h.percentile(50); v != 2 {
		t.Errorf("percentile(50) = %v; want 2", v)
	}
	if v := h.percentile(100); v != 7 {
		t.Errorf("percentile(100) = %v; want 7", v)
	}

	h.inc(10) // goes into overflow bucket
	if v := h.percentile(100); v != 8 {
		t.Errorf("percentile(100) = %v after overflow; want 8", v)
	}
}

func TestHist_Add(t *testing.T) {
	h1 := newHist(10)
	for i := 0; i <= 4; i++ {
		h1.inc(i)
	}
	if v := h1.percentile(50); v != 2 {
		t.Errorf("percentile(50) = %v before add; want 2", v)
	}

	h2 := newHist(10)
	for i := 5; i <= 10; i++ {
		h2.inc(i)
	}
	h1.add(h2)
	if v := h1.percentile(50); v != 5 {
		t.Errorf("percentile(50) = %v after add; want 5", v)
	}
}

// TODO: Write tests for weeklyStats() and averageStats().
