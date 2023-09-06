// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"errors"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/ultravioletrs/agent/agent"
	"github.com/ultravioletrs/manager/manager/qemu"
)

const bootTime = 20 * time.Second

var (
	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("entity not found")
)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	Run(ctx context.Context, computation []byte) (string, error)
}

type managerService struct {
	libvirt *libvirt.Libvirt
	agent   agent.AgentServiceClient
	qemuCfg qemu.Config
}

var _ Service = (*managerService)(nil)

// New instantiates the manager service implementation.
func New(libvirtConn *libvirt.Libvirt, agent agent.AgentServiceClient, qemuCfg qemu.Config) Service {
	return &managerService{
		libvirt: libvirtConn,
		agent:   agent,
		qemuCfg: qemuCfg,
	}
}

func (ms *managerService) Run(ctx context.Context, computation []byte) (string, error) {
	_, err := qemu.CreateVM(ctx, ms.qemuCfg)
	if err != nil {
		return "", err
	}
	// different VM guests can't forward ports to the same ports on the same host
	ms.qemuCfg.HostFwd1++
	ms.qemuCfg.NetDevConfig.HostFwd2++
	ms.qemuCfg.NetDevConfig.HostFwd3++

	time.Sleep(bootTime)

	res, err := ms.agent.Run(ctx, &agent.RunRequest{Computation: computation})
	if err != nil {
		return "", err
	}

	return res.Computation, nil
}
