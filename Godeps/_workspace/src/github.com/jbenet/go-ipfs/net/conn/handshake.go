package conn

import (
	"errors"
	"fmt"

	handshake "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/net/handshake"
	hspb "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/net/handshake/pb"

	context "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/go.net/context"
	proto "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/goprotobuf/proto"
)

// Handshake1 exchanges local and remote versions and compares them
// closes remote and returns an error in case of major difference
func Handshake1(ctx context.Context, c Conn) error {
	rpeer := c.RemotePeer()
	lpeer := c.LocalPeer()

	var remoteH, localH *hspb.Handshake1
	localH = handshake.Handshake1Msg()

	myVerBytes, err := proto.Marshal(localH)
	if err != nil {
		return err
	}

	c.Out() <- myVerBytes
	log.Debugf("Sent my version (%s) to %s", localH, rpeer)

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-c.Closing():
		return errors.New("remote closed connection during version exchange")

	case data, ok := <-c.In():
		if !ok {
			return fmt.Errorf("error retrieving from conn: %v", rpeer)
		}

		remoteH = new(hspb.Handshake1)
		err = proto.Unmarshal(data, remoteH)
		if err != nil {
			return fmt.Errorf("could not decode remote version: %q", err)
		}

		log.Debugf("Received remote version (%s) from %s", remoteH, rpeer)
	}

	if err := handshake.Handshake1Compatible(localH, remoteH); err != nil {
		log.Infof("%s (%s) incompatible version with %s (%s)", lpeer, localH, rpeer, remoteH)
		return err
	}

	log.Debugf("%s version handshake compatible %s", lpeer, rpeer)
	return nil
}

// Handshake3 exchanges local and remote service information
func Handshake3(ctx context.Context, c Conn) (*handshake.Handshake3Result, error) {
	rpeer := c.RemotePeer()
	lpeer := c.LocalPeer()

	// setup + send the message to remote
	var remoteH, localH *hspb.Handshake3
	localH = handshake.Handshake3Msg(lpeer, c.RemoteMultiaddr())
	localB, err := proto.Marshal(localH)
	if err != nil {
		return nil, err
	}

	c.Out() <- localB
	log.Debugf("Handshake1: sent to %s", rpeer)

	// wait + listen for response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case <-c.Closing():
		return nil, errors.New("Handshake3: error remote connection closed")

	case remoteB, ok := <-c.In():
		if !ok {
			return nil, fmt.Errorf("Handshake3 error receiving from conn: %v", rpeer)
		}

		remoteH = new(hspb.Handshake3)
		err = proto.Unmarshal(remoteB, remoteH)
		if err != nil {
			return nil, fmt.Errorf("Handshake3 could not decode remote msg: %q", err)
		}

		log.Debugf("Handshake3 received from %s", rpeer)
	}

	// actually update our state based on the new knowledge
	res, err := handshake.Handshake3Update(lpeer, rpeer, remoteH)
	if err != nil {
		log.Errorf("Handshake3 failed to update %s", rpeer)
	}
	res.RemoteObservedAddress = c.RemoteMultiaddr()
	return res, nil
}
