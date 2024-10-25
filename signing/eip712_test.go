package signing

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

type Message struct {
	A    uint64      `eip712:"a"`
	C    *big.Int    `eip712:"c"`
	Msg1 *Message2   `eip712:"msg1"`
	Msg2 []Message2  `eip712:"msg2"`
	Msg3 [3]Message2 `eip712:"msg3"`
}

type Message2 struct {
	B    uint64        `eip712:"b"`
	Addr ethgo.Address `eip712:"addr"`
}

func TestBuildMessage_Encode(t *testing.T) {
	domain := &EIP712Domain{
		Name: "name1",
	}

	b := NewEIP712MessageBuilder[Message](domain)
	require.Equal(t, "Message(uint64 a,uint256 c,Message2 msg1,Message2[] msg2,Message2[3] msg3)Message2(uint64 b,address addr)", b.GetEncodedType())

	msg := &Message{
		C:    big.NewInt(1),
		Msg1: &Message2{},
		Msg2: []Message2{
			{B: 1},
		},
	}
	typedMsg := b.Build(msg)

	_, ok := typedMsg.Message["msg1"].(interface{})
	require.True(t, ok)

	_, ok = typedMsg.Message["msg2"].([]interface{})
	require.True(t, ok)

	_, ok = typedMsg.Message["msg3"].([3]interface{})
	require.True(t, ok)

	_, err := typedMsg.Hash()
	require.NoError(t, err)
}

func TestBuildMessage_BasicTypes(t *testing.T) {
	domain := &EIP712Domain{
		Name: "name1",
	}

	type Message struct {
		A uint64
		B uint32
		C uint16
		D uint8
		E [32]byte
		F string
	}

	b := NewEIP712MessageBuilder[Message](domain)
	require.Equal(t, "Message(uint64 A,uint32 B,uint16 C,uint8 D,[32]byte E,string F)", b.GetEncodedType())
}
