package connmgr

import (
	"context"

	"github.com/RTradeLtd/libp2px-core/network"
	"github.com/RTradeLtd/libp2px-core/peer"
)

// NullConnMgr is a ConnMgr that provides no functionality.
type NullConnMgr struct{}

var _ ConnManager = (*NullConnMgr)(nil)

func (ncm NullConnMgr) TagPeer(peer.ID, string, int)             {}
func (ncm NullConnMgr) UntagPeer(peer.ID, string)                {}
func (ncm NullConnMgr) UpsertTag(peer.ID, string, func(int) int) {}
func (ncm NullConnMgr) GetTagInfo(peer.ID) *TagInfo              { return &TagInfo{} }
func (ncm NullConnMgr) TrimOpenConns(ctx context.Context)        {}
func (ncm NullConnMgr) Notifee() network.Notifiee                { return network.GlobalNoopNotifiee }
func (ncm NullConnMgr) Protect(peer.ID, string)                  {}
func (ncm NullConnMgr) Unprotect(peer.ID, string) bool           { return false }
func (ncm NullConnMgr) Close() error                             { return nil }
