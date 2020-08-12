// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import "testing"

func TestStats_Update(t *testing.T) {
	s := newStats()
	s.update(molecular, positive, age20To29, 1)
	s.update(molecular, negative, age20To29, 2)
	s.update(molecular, positive, age30To39, 3)
	s.update(molecular, otherResult, age40To49, 3)
	s.update(molecular, negative, age0To9, 4)
	s.update(molecular, positive, age60To69, 5)
	s.update(serological, positive, age10To19, 6)
	s.update(antigen, positive, age50To59, 7)
	s.update(antigen, negative, age80To89, 8)

	if s.pos != 3 {
		t.Errorf("pos = %v; want 3", s.pos)
	}
	if s.neg != 2 {
		t.Errorf("neg = %v; want 2", s.neg)
	}
	if s.other != 1 {
		t.Errorf("other = %v; want 1", s.other)
	}
	if s.ab != 1 {
		t.Errorf("ab = %v; want 1", s.ab)
	}
	if s.ag != 2 {
		t.Errorf("ag = %v; want 2", s.ag)
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
	if v := s.posDelayPct(100); v != 5 {
		t.Errorf("posDelayPct(100) = %v; want 5", v)
	}
	if v := s.negDelayPct(100); v != 4 {
		t.Errorf("negDelayPct(100) = %v; want 4", v)
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
