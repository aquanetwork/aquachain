// +build !usb

package usbwallet

import (
	"errors"

	"gitlab.com/aquachain/aquachain/aqua/accounts"
	"gitlab.com/aquachain/aquachain/aqua/event"
)

// Hub is a accounts.Backend that can find and handle generic USB hardware wallets.
type Hub struct{}

// NewLedgerHub creates a new hardware wallet manager for Ledger devices.
func NewLedgerHub() (*Hub, error) {
	return nil, accounts.ErrNotSupported
}

// ErrTrezorPINNeeded is returned if opening the trezor requires a PIN code. In
// this case, the calling application should display a pinpad and send back the
// encoded passphrase.
var ErrTrezorPINNeeded = errors.New("trezor: pin needed")

// NewTrezorHub creates a new hardware wallet manager for Trezor devices.
func NewTrezorHub() (*Hub, error) {
	return nil, accounts.ErrNotSupported
}

// Wallets implements accounts.Backend, returning all the currently tracked USB
// devices that appear to be hardware wallets.
func (hub *Hub) Wallets() []accounts.Wallet {
	return nil
}

// Subscribe implements accounts.Backend, creating an async subscription to
// receive notifications on the addition or removal of USB wallets.
func (hub *Hub) Subscribe(sink chan<- accounts.WalletEvent) event.Subscription {
	return nil
}
