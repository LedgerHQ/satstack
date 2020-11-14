package protocol

import "errors"

var (
	// ErrDecodeHex indicates the transaction hex could not be decoded.
	ErrDecodeHex = errors.New("failed to decode hex")

	// ErrMsgTxDeserialize indicates that the parser could not process the
	// serialized hex to wire.MsgTx.
	ErrMsgTxDeserialize = errors.New("failed to deserialize to MsgTx")
)
