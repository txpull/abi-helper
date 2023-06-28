package solidity

import "context"

type Decoder struct {
	ctx       context.Context
	rawSource string
}

func NewDecoder(ctx context.Context, rawSource string) *Decoder {
	return &Decoder{
		ctx:       ctx,
		rawSource: rawSource,
	}
}
