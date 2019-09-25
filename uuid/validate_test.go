package uuid

import (
	"strings"
	"testing"
)

func TestValidAllowed(t *testing.T) {

	valid := []string{
		"df437eb6-dfa5-11e9-8a34-2a2ae2dbcce4",
		"ece98b14-dfa5-11e9-8a34-2a2ae2dbcce4",
		"f776b3a4-dfa5-11e9-8a34-2a2ae2dbcce4",
		"db825f44-4404-4fc4-aa27-0d1beb109350",
		"8b1ec84f-4a2b-4973-a7d7-e70f90c86540",
		"49aed468-e271-4984-b8e9-4170912e5b2f",
	}

	for _, u := range valid {

		if !ValidFormat(u) {
			t.Fatalf("Incorrectly identified as invalid: %s", u)
		}

		uu := strings.ToUpper(u)

		if !ValidFormat(uu) {
			t.Fatalf("Incorrectly identified as invalid: %s", uu)
		}

	}

}

func TestInvalidRejected(t *testing.T) {

	valid := []string{
		"df437eb6-dfa5-11e9-8a34-2a2adbcce4",
		"ece98b14-dfa5-11e9-8g34-2a2ae2dbcce4",
		"f776b3a4-dfa5-11e9a-34-2a2ae2dbcce4",
		"db825f4-44404-4fc4-aa27-0d1beb109350",
		"8b1ec84f-4a2-44973-a7d7-e70f90c86540",
		"49aed468-e271-498-4b8e9-4170912e5b2f",
		"",
		"-----",
	}

	for _, u := range valid {

		if ValidFormat(u) {
			t.Fatalf("Incorrectly identified as valid: %s", u)
		}

		uu := strings.ToUpper(u)

		if ValidFormat(uu) {
			t.Fatalf("Incorrectly identified as valid: %s", uu)
		}

	}

}

func TestFieldExtraction(t *testing.T) {

	/*
	   UUID                   = time-low "-" time-mid "-"
	                            time-high-and-version "-"
	                            clock-seq-and-reserved
	                            clock-seq-low "-" node
	   time-low               = 4hexOctet
	   time-mid               = 2hexOctet
	   time-high-and-version  = 2hexOctet
	   clock-seq-and-reserved = hexOctet
	   clock-seq-low          = hexOctet
	   node                   = 6hexOctet
	   hexOctet               = hexDigit hexDigit
	*/

	uuid := "92e8a690-98e3-424b-83e4-cecb3548a7a0"

	if extractField(timeLow, uuid) != "92e8a690" {
		t.Errorf("Could not extract time-low  ")
	}

	if extractField(timeMid, uuid) != "98e3" {
		t.Errorf("Could not extract time-mid  ")
	}

	if extractField(timeHighAndVersion, uuid) != "424b" {
		t.Errorf("Could not extract time-high-and-version")
	}

	if extractField(clockSeqAndReserved, uuid) != "83" {
		t.Errorf("Could not extract clock-seq-and-reserved")
	}

	if extractField(clockSeqLow, uuid) != "e4" {
		t.Errorf("Could not extract clock-seq-low")
	}

	if extractField(node, uuid) != "cecb3548a7a0" {
		t.Errorf("Could not extract node")
	}
}

func TestBinaryExtraction(t *testing.T) {

	uuid := "92e8a690-98e3-424b-83e4-cecb3548a7a0"

	if extractBinaryField(timeLow, uuid) != "10010010111010001010011010010000" {
		t.Errorf("Could not extract time-low  ")
	}

	if extractBinaryField(timeMid, uuid) != "1001100011100011" {
		t.Errorf("Could not extract time-mid  ")
	}

	thv := extractBinaryField(timeHighAndVersion, uuid)

	if thv != "0100001001001011" {
		t.Errorf("Could not extract time-high-and-version, was %s", thv)
	}

	if extractBinaryField(clockSeqAndReserved, uuid) != "10000011" {
		t.Errorf("Could not extract clock-seq-and-reserved")
	}

	if extractBinaryField(clockSeqLow, uuid) != "11100100" {
		t.Errorf("Could not extract clock-seq-low")
	}

	if extractBinaryField(node, uuid) != "110011101100101100110101010010001010011110100000" {
		t.Errorf("Could not extract node")
	}
}
