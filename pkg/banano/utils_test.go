package banano

import "testing"

func TestGenPublicKey(t *testing.T) {

	var pkTest = []struct {
		acc      string // input
		expected string // expected result
	}{
		{"ban_1srfj9ohtnh9zkwuxdscf8n7dqcndks9waurjcorj6aizwagckampump8s4k", "670D89EAFD51E7FCB9BEAF2A69A855DD545CB27E23788AAB889110FF10E54913"},
		{"ban_3aefzmig9tkfjhrtu88pr4cmwh1cm7fk6btgr3g67jyfihfwwhpkf16dpqrg", "A18DFCE0E3EA4D8BF1AD98D6C0953E3C0A995B22274EC05C42C7CD83DBCE3ED2"},
		{"ban_1111111111111111111111111111111111111111111111111111hifc8npp", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"ban_1w589t1rkfgab1tq8mhu54xarp151angcs6jd8byhycy5suyqzhtk4j699fd", "70663E818935C84835734DFB18BA8C58030228E564915993E7F95E1E77EBFDFA"},
	}

	for _, tt := range pkTest {
		actual, _ := GenPublicKey(tt.acc)
		if actual != tt.expected {
			t.Errorf("failed public key generation: expected %s, actual %s", tt.expected, actual)
		}
	}
}

func TestYellowSpyGlassOpened(t *testing.T) {
	var tests = []struct {
		acc      string // input
		expected string // expected result
	}{
		{"ban_1srfj9ohtnh9zkwuxdscf8n7dqcndks9waurjcorj6aizwagckampump8s4k", "670D89EAFD51E7FCB9BEAF2A69A855DD545CB27E23788AAB889110FF10E54913"},
		{"ban_3aefzmig9tkfjhrtu88pr4cmwh1cm7fk6btgr3g67jyfihfwwhpkf16dpqrg", "A18DFCE0E3EA4D8BF1AD98D6C0953E3C0A995B22274EC05C42C7CD83DBCE3ED2"},
		{"ban_1111111111111111111111111111111111111111111111111111hifc8npp", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"ban_1w589t1rkfgab1tq8mhu54xarp151angcs6jd8byhycy5suyqzhtk4j699fd", "70663E818935C84835734DFB18BA8C58030228E564915993E7F95E1E77EBFDFA"},
	}

	for _, tt := range tests {
		actual, _ := GenPublicKey(tt.acc)
		if actual != tt.expected {
			t.Errorf("failed public key generation: expected %s, actual %s", tt.expected, actual)
		}
	}
}
