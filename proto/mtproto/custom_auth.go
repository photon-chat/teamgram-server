/*
 * Custom auth TL codec, class ID registration, and RPC routing
 */

package mtproto

import (
	"github.com/gogo/protobuf/jsonpb"
)

///////////////////////////////////////////////////////////////////////////////
// TL Encode/Decode for custom auth types
///////////////////////////////////////////////////////////////////////////////

// Auth_AuthMethods
func (m *Auth_AuthMethods) Encode(x *EncodeBuf, layer int32) error {
	x.UInt(0x33323197) // CRC32_auth_authMethods
	x.VectorString(m.GetMethods())
	return nil
}

func (m *Auth_AuthMethods) CalcByteSize(layer int32) int { return 0 }

func (m *Auth_AuthMethods) Decode(dBuf *DecodeBuf) error {
	m.Methods = dBuf.VectorString()
	return dBuf.GetError()
}

func (m *Auth_AuthMethods) DebugString() string {
	jsonm := &jsonpb.Marshaler{OrigName: true}
	s, _ := jsonm.MarshalToString(m)
	return s
}

// TLAuthGetAuthMethods
func (m *TLAuthGetAuthMethods) Encode(x *EncodeBuf, layer int32) error {
	x.UInt(0x03f8bdbb) // CRC32_auth_getAuthMethods
	return nil
}

func (m *TLAuthGetAuthMethods) CalcByteSize(layer int32) int { return 0 }

func (m *TLAuthGetAuthMethods) Decode(dBuf *DecodeBuf) error {
	return dBuf.GetError()
}

func (m *TLAuthGetAuthMethods) DebugString() string {
	jsonm := &jsonpb.Marshaler{OrigName: true}
	s, _ := jsonm.MarshalToString(m)
	return s
}

// TLAuthUsernameRegister
func (m *TLAuthUsernameRegister) Encode(x *EncodeBuf, layer int32) error {
	x.UInt(0x96b51340) // CRC32_auth_usernameRegister
	x.String(m.GetUsername())
	x.String(m.GetPassword())
	x.String(m.GetFirstName())
	return nil
}

func (m *TLAuthUsernameRegister) CalcByteSize(layer int32) int { return 0 }

func (m *TLAuthUsernameRegister) Decode(dBuf *DecodeBuf) error {
	m.Username = dBuf.String()
	m.Password = dBuf.String()
	m.FirstName = dBuf.String()
	return dBuf.GetError()
}

func (m *TLAuthUsernameRegister) DebugString() string {
	jsonm := &jsonpb.Marshaler{OrigName: true}
	s, _ := jsonm.MarshalToString(m)
	return s
}

// TLAuthUsernameSignIn
func (m *TLAuthUsernameSignIn) Encode(x *EncodeBuf, layer int32) error {
	x.UInt(0x6b864f99) // CRC32_auth_usernameSignIn
	x.String(m.GetUsername())
	x.String(m.GetPassword())
	return nil
}

func (m *TLAuthUsernameSignIn) CalcByteSize(layer int32) int { return 0 }

func (m *TLAuthUsernameSignIn) Decode(dBuf *DecodeBuf) error {
	m.Username = dBuf.String()
	m.Password = dBuf.String()
	return dBuf.GetError()
}

func (m *TLAuthUsernameSignIn) DebugString() string {
	jsonm := &jsonpb.Marshaler{OrigName: true}
	s, _ := jsonm.MarshalToString(m)
	return s
}

// TLAuthPhonePasswordRegister
func (m *TLAuthPhonePasswordRegister) Encode(x *EncodeBuf, layer int32) error {
	x.UInt(0x4e602932) // CRC32_auth_phonePasswordRegister
	x.String(m.GetPhone())
	x.String(m.GetPassword())
	x.String(m.GetFirstName())
	return nil
}

func (m *TLAuthPhonePasswordRegister) CalcByteSize(layer int32) int { return 0 }

func (m *TLAuthPhonePasswordRegister) Decode(dBuf *DecodeBuf) error {
	m.Phone = dBuf.String()
	m.Password = dBuf.String()
	m.FirstName = dBuf.String()
	return dBuf.GetError()
}

func (m *TLAuthPhonePasswordRegister) DebugString() string {
	jsonm := &jsonpb.Marshaler{OrigName: true}
	s, _ := jsonm.MarshalToString(m)
	return s
}

// TLAuthPhonePasswordSignIn
func (m *TLAuthPhonePasswordSignIn) Encode(x *EncodeBuf, layer int32) error {
	x.UInt(0x77053d31) // CRC32_auth_phonePasswordSignIn
	x.String(m.GetPhone())
	x.String(m.GetPassword())
	return nil
}

func (m *TLAuthPhonePasswordSignIn) CalcByteSize(layer int32) int { return 0 }

func (m *TLAuthPhonePasswordSignIn) Decode(dBuf *DecodeBuf) error {
	m.Phone = dBuf.String()
	m.Password = dBuf.String()
	return dBuf.GetError()
}

func (m *TLAuthPhonePasswordSignIn) DebugString() string {
	jsonm := &jsonpb.Marshaler{OrigName: true}
	s, _ := jsonm.MarshalToString(m)
	return s
}

///////////////////////////////////////////////////////////////////////////////
// Registration
///////////////////////////////////////////////////////////////////////////////

func init() {
	// Class ID registration (constructor ID -> struct factory)
	clazzIdRegisters2[int32(CRC32_auth_authMethods)] = func() TLObject {
		return &Auth_AuthMethods{Constructor: CRC32_auth_authMethods}
	}
	clazzIdRegisters2[int32(CRC32_auth_getAuthMethods)] = func() TLObject {
		return &TLAuthGetAuthMethods{Constructor: CRC32_auth_getAuthMethods}
	}
	clazzIdRegisters2[int32(CRC32_auth_usernameRegister)] = func() TLObject {
		return &TLAuthUsernameRegister{Constructor: CRC32_auth_usernameRegister}
	}
	clazzIdRegisters2[int32(CRC32_auth_usernameSignIn)] = func() TLObject {
		return &TLAuthUsernameSignIn{Constructor: CRC32_auth_usernameSignIn}
	}
	clazzIdRegisters2[int32(CRC32_auth_phonePasswordRegister)] = func() TLObject {
		return &TLAuthPhonePasswordRegister{Constructor: CRC32_auth_phonePasswordRegister}
	}
	clazzIdRegisters2[int32(CRC32_auth_phonePasswordSignIn)] = func() TLObject {
		return &TLAuthPhonePasswordSignIn{Constructor: CRC32_auth_phonePasswordSignIn}
	}

	// RPC routing registration (struct name -> gRPC method path)
	rpcContextRegisters["TLAuthGetAuthMethods"] = RPCContextTuple{"/mtproto.RPCAuthorization/auth_getAuthMethods", func() interface{} { return new(Auth_AuthMethods) }}
	rpcContextRegisters["TLAuthUsernameRegister"] = RPCContextTuple{"/mtproto.RPCAuthorization/auth_usernameRegister", func() interface{} { return new(Auth_Authorization) }}
	rpcContextRegisters["TLAuthUsernameSignIn"] = RPCContextTuple{"/mtproto.RPCAuthorization/auth_usernameSignIn", func() interface{} { return new(Auth_Authorization) }}
	rpcContextRegisters["TLAuthPhonePasswordRegister"] = RPCContextTuple{"/mtproto.RPCAuthorization/auth_phonePasswordRegister", func() interface{} { return new(Auth_Authorization) }}
	rpcContextRegisters["TLAuthPhonePasswordSignIn"] = RPCContextTuple{"/mtproto.RPCAuthorization/auth_phonePasswordSignIn", func() interface{} { return new(Auth_Authorization) }}
}
