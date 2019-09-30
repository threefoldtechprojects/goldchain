package types

import "testing"

func TestCustodyFeeUnlockHashString(t *testing.T) {
	s := CustodyFeeUnlockHash.String()
	if s != "800000000000000000000000000000000000000000000000000000000000000000af7bedde1fea" {
		t.Error(s, "!=", "800000000000000000000000000000000000000000000000000000000000000000af7bedde1fea")
	}
}
