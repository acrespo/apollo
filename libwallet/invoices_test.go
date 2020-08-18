package libwallet

import (
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/zpay32"
)

func TestInvoiceSecrets(t *testing.T) {
	setup()

	network := Regtest()

	userKey, _ := NewHDPrivateKey(randomBytes(32), network)
	userKey.Path = "m/0'/0'"
	muunKey, _ := NewHDPrivateKey(randomBytes(32), network)
	muunKey.Path = "m/0'/0'"

	secrets, err := GenerateInvoiceSecrets(userKey.PublicKey(), muunKey.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 5 {
		t.Fatalf("expected 5 new secrets, got %d", len(secrets))
	}

	err = PersistInvoiceSecrets(secrets)
	if err != nil {
		t.Fatal(err)
	}

	// try to generate more secrets
	moreSecrets, err := GenerateInvoiceSecrets(userKey.PublicKey(), muunKey.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	if len(moreSecrets) != 0 {
		t.Fatal("expected no new secrets to be created")
	}

	routeHints := &RouteHints{
		Pubkey:                    "03c48d1ff96fa32e2776f71bba02102ffc2a1b91e2136586418607d32e762869fd",
		FeeBaseMsat:               1000,
		FeeProportionalMillionths: 1000,
		CltvExpiryDelta:           8,
	}

	invoice, err := CreateInvoice(network, userKey, routeHints, &InvoiceOptions{
		AmountSat:   1000,
		Description: "hello world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if invoice == "" {
		t.Fatal("expected non-empty invoice string")
	}

	payreq, err := zpay32.Decode(invoice, network.network)
	if err != nil {
		t.Fatal(err)
	}
	if !payreq.Features.HasFeature(lnwire.TLVOnionPayloadOptional) {
		t.Fatal("expected invoice to have var onion optin feature")
	}
	if !payreq.Features.HasFeature(lnwire.PaymentAddrOptional) {
		t.Fatal("expected invoice to have payment secret feature")
	}
	if payreq.MilliSat.ToSatoshis() != btcutil.Amount(1000) {
		t.Fatalf("expected invoice amount to be 1000 sats, got %v", payreq.MilliSat)
	}
	if payreq.Description == nil || *payreq.Description != "hello world" {
		t.Fatalf("expected payment description to match, got %v", payreq.Description)
	}
	if payreq.MinFinalCLTVExpiry() != 144 {
		t.Fatalf("expected min final CLTV expiry to be 144, got %v", payreq.MinFinalCLTVExpiry())
	}
	if len(payreq.RouteHints) == 0 {
		t.Fatalf("expected invoice to contain route hints")
	}
	hopHints := payreq.RouteHints[0]
	if len(hopHints) != 1 {
		t.Fatalf("expected invoice route hints to contain exactly 1 hop hint")
	}
	if hopHints[0].ChannelID&(1<<63) == 0 {
		t.Fatal("invalid short channel id in hophints")
	}
	if hopHints[0].FeeBaseMSat != 1000 {
		t.Fatalf("expected fee base to be 1000 msat, got %v instead", hopHints[0].FeeBaseMSat)
	}
	if hopHints[0].FeeProportionalMillionths != 1000 {
		t.Fatalf("expected fee proportional millionths to be 1000, got %v instead", hopHints[0].FeeProportionalMillionths)
	}
	if hopHints[0].CLTVExpiryDelta != 8 {
		t.Fatalf("expected CLTV expiry delta to be 8, got %v instead", hopHints[0].CLTVExpiryDelta)
	}

	invoice2, err := CreateInvoice(network, userKey, routeHints, &InvoiceOptions{})
	if err != nil {
		t.Fatal(err)
	}

	payreq2, err := zpay32.Decode(invoice2, network.network)
	if err != nil {
		t.Fatal(err)
	}

	if payreq.PaymentHash == payreq2.PaymentHash {
		t.Fatal("successive invoice payment hashes should be different")
	}

}
