package logging

import "testing"

func TestStringToNumericLevel(t *testing.T) {
	if l, err := LogLevelFromLabel(AllLabel); err != nil || l != All {
		t.Fail()
	}

	if l, err := LogLevelFromLabel(TraceLabel); err != nil || l != Trace {
		t.Fail()
	}

	if l, err := LogLevelFromLabel(DebugLabel); err != nil || l != Debug {
		t.Fail()
	}

	if l, err := LogLevelFromLabel(WarnLabel); err != nil || l != Warn {
		t.Fail()
	}

	if l, err := LogLevelFromLabel(ErrorLabel); err != nil || l != Error {
		t.Fail()
	}

	if l, err := LogLevelFromLabel(InfoLabel); err != nil || l != Info {
		t.Fail()
	}

	if l, err := LogLevelFromLabel(FatalLabel); err != nil || l != Fatal {
		t.Fail()
	}

	if _, err := LogLevelFromLabel("XXXXX"); err == nil {
		t.Fail()
	}
}

func TestNumericToStringLevel(t *testing.T) {

	if m := LabelFromLevel(All); m != AllLabel {
		t.Fail()
	}

	if m := LabelFromLevel(Trace); m != TraceLabel {
		t.Fail()
	}

	if m := LabelFromLevel(Debug); m != DebugLabel {
		t.Fail()
	}

	if m := LabelFromLevel(Warn); m != WarnLabel {
		t.Fail()
	}

	if m := LabelFromLevel(Error); m != ErrorLabel {
		t.Fail()
	}

	if m := LabelFromLevel(Info); m != InfoLabel {
		t.Fail()
	}

	if m := LabelFromLevel(Fatal); m != FatalLabel {
		t.Fail()
	}

	c := LogLevel(999)

	if m := LabelFromLevel(c); m != "CUSTOM" {
		t.Fail()
	}
}
