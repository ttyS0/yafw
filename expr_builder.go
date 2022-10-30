package yafw

import (
	"encoding/binary"
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

type ExprBuilder struct {
	expr []expr.Any
}

func (eb *ExprBuilder) Exprs() []expr.Any {
	return eb.expr
}

func (eb *ExprBuilder) MetaEgressInterface(register uint32) *ExprBuilder {
	return eb.Append(
		&expr.Meta{
			Key:      expr.MetaKeyOIFNAME,
			Register: register,
		},
	)
}

func (eb *ExprBuilder) MetaIngressInterface(register uint32) *ExprBuilder {
	return eb.Append(
		&expr.Meta{
			Key:      expr.MetaKeyIIFNAME,
			Register: register,
		},
	)
}

func (eb *ExprBuilder) MetaL4Protocol(register uint32) *ExprBuilder {
	return eb.Append(
		&expr.Meta{
			Key:      expr.MetaKeyL4PROTO,
			Register: register,
		},
	)
}

func (eb *ExprBuilder) CompareL4Protocol(register uint32, protocol uint8) *ExprBuilder {
	return eb.Append(
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: register,
			Data:     []byte{protocol},
		},
	)
}

func (eb *ExprBuilder) LoadSourcePort(register uint32) *ExprBuilder {
	return eb.Append(
		&expr.Payload{
			DestRegister: register,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       0,
			Len:          2,
		},
	)
}

func (eb *ExprBuilder) LoadDestinationPort(register uint32) *ExprBuilder {
	return eb.Append(
		&expr.Payload{
			DestRegister: register,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       2,
			Len:          2,
		},
	)
}

func (eb *ExprBuilder) ComparePort(register uint32, port uint16) *ExprBuilder {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, port)
	return eb.Append(
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     data,
		},
	)
}

func (eb *ExprBuilder) ComparePortRange(register uint32, min, max uint16) *ExprBuilder {

	from := make([]byte, 2)
	binary.BigEndian.PutUint16(from, min)
	to := make([]byte, 2)
	binary.BigEndian.PutUint16(to, max)
	return eb.Append(
		&expr.Range{
			Op:       expr.CmpOpEq,
			Register: 1,
			FromData: from,
			ToData:   to,
		},
	)
}

func (eb *ExprBuilder) PayloadIPSource(register uint32) *ExprBuilder {
	return eb.Append(
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       12,
			Len:          4,
		},
	)
}

func (eb *ExprBuilder) PayloadIPDestination(register uint32) *ExprBuilder {
	return eb.Append(
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       16,
			Len:          4,
		},
	)
}

func (eb *ExprBuilder) CompareInterfaceName(register uint32, name string) *ExprBuilder {
	return eb.Append(
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: register,
			Data:     InterfaceName(name),
		},
	)
}

func (eb *ExprBuilder) CompareIPRange(register uint32, iprange *IPRange) *ExprBuilder {
	return eb.Append(
		&expr.Cmp{
			Op:       expr.CmpOpGte,
			Register: register,
			Data:     iprange.First().To4(),
		},
		&expr.Cmp{
			Op:       expr.CmpOpLt,
			Register: register,
			Data:     iprange.End().To4(),
		},
	)
}

func (eb *ExprBuilder) LookupZone(register uint32, zone *Zone) *ExprBuilder {
	return eb.Append(
		&expr.Lookup{
			SourceRegister: register,
			SetID:          zone.set.ID,
			SetName:        zone.set.Name,
		},
	)
}

func (eb *ExprBuilder) LookupSet(register uint32, set *nftables.Set) *ExprBuilder {
	if set.Anonymous {
		return eb.Append(
			&expr.Lookup{
				SourceRegister: register,
				SetName:        set.Name,
				SetID:          set.ID,
			},
		)
	} else {
		return eb.Append(
			&expr.Lookup{
				SourceRegister: register,
				SetName:        set.Name,
				SetID:          set.ID,
			},
		)
	}
}

func (eb *ExprBuilder) Masquerade() *ExprBuilder {
	return eb.Append(&expr.Masq{})
}

func (eb *ExprBuilder) SourceNATIP(first net.IP, last net.IP) *ExprBuilder {
	return eb.Append(
		&expr.Immediate{
			Register: 1,
			Data:     first.To4(),
		},
		&expr.Immediate{
			Register: 2,
			Data:     last.To4(),
		},
		&expr.NAT{
			Type:        expr.NATTypeSourceNAT,
			Family:      unix.NFPROTO_IPV4,
			RegAddrMin:  1,
			RegProtoMax: 2,
		},
	)
}

func (eb *ExprBuilder) SourceNATIPRange(start net.IP, end net.IP) *ExprBuilder {
	return eb.Append(
		&expr.Immediate{
			Register: 1,
			Data:     start.To4(),
		},
		&expr.Immediate{
			Register: 2,
			Data:     end.To4(),
		},
		&expr.NAT{
			Type:       expr.NATTypeSourceNAT,
			Family:     unix.NFPROTO_IPV4,
			RegAddrMin: 1,
			RegAddrMax: 2,
		},
	)
}

func (eb *ExprBuilder) ConntrackState(state uint32) *ExprBuilder {
	return eb.Append(
		&expr.Ct{Register: 1, SourceRegister: false, Key: expr.CtKeySTATE},
		&expr.Bitwise{
			SourceRegister: 1,
			DestRegister:   1,
			Len:            4,
			Mask:           binaryutil.NativeEndian.PutUint32(state),
			Xor:            binaryutil.NativeEndian.PutUint32(0),
		},
		&expr.Cmp{Op: expr.CmpOpNeq, Register: 1, Data: binaryutil.NativeEndian.PutUint32(0)},
	)
}

func (eb *ExprBuilder) VerdictDrop() *ExprBuilder {
	return eb.Append(
		&expr.Verdict{
			Kind: expr.VerdictDrop,
		},
	)
}

func (eb *ExprBuilder) VerdictAccept() *ExprBuilder {
	return eb.Append(
		&expr.Verdict{
			Kind: expr.VerdictAccept,
		},
	)
}

func (eb *ExprBuilder) LogIPOptions(prefix string) *ExprBuilder {
	return eb.Append(
		&expr.Log{
			Data:  []byte(prefix),
			Flags: expr.LogFlagsIPOpt,
		},
	)
}

func (eb *ExprBuilder) Counter() *ExprBuilder {
	return eb.Append(
		&expr.Counter{},
	)
}

func (eb *ExprBuilder) Append(args ...expr.Any) *ExprBuilder {
	eb.expr = append(eb.expr, args...)

	return eb
}

func (eb *ExprBuilder) AppendGroup(args ...[]expr.Any) *ExprBuilder {
	for _, group := range args {
		eb.expr = append(eb.expr, group...)
	}

	return eb
}
