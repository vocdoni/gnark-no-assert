package ecdsa

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"
)

// Signature represents the signature for some message.
type Signature[Scalar emulated.FieldParams] struct {
	R, S emulated.Element[Scalar]
}

// PublicKey represents the public key to verify the signature for.
type PublicKey[Base, Scalar emulated.FieldParams] sw_emulated.AffinePoint[Base]

// Verify asserts that the signature sig verifies for the message msg and public
// key pk. The curve parameters params define the elliptic curve.
//
// We assume that the message msg is already hashed to the scalar field.
func (pk PublicKey[T, S]) Verify(api frontend.API, params sw_emulated.CurveParams, msg *emulated.Element[S], sig *Signature[S]) {
	flag := pk.SignIsValid(api, params, msg, sig)
	api.AssertIsEqual(flag, 1)
}

// SignIsValid returns 1 if the signature sig verifies for the message msg and
// public key pk or 0 if not. The curve parameters params define the elliptic
// curve.
//
// We assume that the message msg is already hashed to the scalar field.
func (pk PublicKey[T, S]) SignIsValid(api frontend.API, params sw_emulated.CurveParams, msg *emulated.Element[S], sig *Signature[S]) frontend.Variable {
	cr, err := sw_emulated.New[T, S](api, params)
	if err != nil {
		panic(err)
	}
	scalarApi, err := emulated.NewField[S](api)
	if err != nil {
		panic(err)
	}
	baseApi, err := emulated.NewField[T](api)
	if err != nil {
		panic(err)
	}
	pkpt := sw_emulated.AffinePoint[T](pk)
	msInv := scalarApi.Div(msg, &sig.S)
	rsInv := scalarApi.Div(&sig.R, &sig.S)

	// q = [rsInv]pkpt + [msInv]g
	q := cr.JointScalarMulBase(&pkpt, rsInv, msInv)
	qx := baseApi.Reduce(&q.X)
	qxBits := baseApi.ToBits(qx)
	rbits := scalarApi.ToBits(&sig.R)
	if len(rbits) != len(qxBits) {
		panic("non-equal lengths")
	}
	// store 1 to expect equality
	res := frontend.Variable(1)
	for i := range rbits {
		// calc the difference between the bits
		diff := api.Sub(rbits[i], qxBits[i])
		// update the result with the AND of the previous result and the
		// equality between the bits (diff == 0)
		res = api.And(res, api.IsZero(diff))
	}
	return res
}
