package grpcutils

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/metadata"
)

type ctxType int

const (
	requestCtxKey ctxType = iota
)

const (
	deviceIDMDKey     = "device-id"
	sessionIDMDKey    = "session-id"
	requestIDMDKey    = "x-request-id"
	forwardedForMDKey = "x-forwarded-for"
	userAgentMDKey    = "user-agent"
)

type RequestMetaData struct {
	SessionID string
	RequestID string
	DeviceID  string
	RealIP    string
	ProxyIPs  []string
	UserAgent string
}

// GetRequestMetaData returns RequestMetaData from the context.
func GetRequestMetaData(ctx context.Context) *RequestMetaData {
	return ctx.Value(requestCtxKey).(*RequestMetaData)
}

// SetRequestMetaData returns a new context with set RequestMetaData.
func SetRequestMetaData(ctx context.Context, s *RequestMetaData) context.Context {
	return context.WithValue(ctx, requestCtxKey, s)
}

// ParseRequestMetaData returns request meta data from context MetaData gRPC.
func ParseRequestMetaData(ctx context.Context) (md *RequestMetaData, err error) {
	md = &RequestMetaData{}

	md.RealIP, md.ProxyIPs, err = ForwardedFor(ctx)
	if err != nil {
		return nil, err
	}

	md.UserAgent, err = UserAgent(ctx)
	if err != nil {
		return nil, err
	}

	md.DeviceID, err = DeviceID(ctx)
	if err != nil {
		return nil, err
	}

	md.RequestID, err = RequestID(ctx)
	if err != nil {
		return nil, err
	}

	md.SessionID, _ = SessionID(ctx)

	return md, nil
}

// ForwardedFor returns real IP and proxy IPs from context gRPC MetaData.
func ForwardedFor(ctx context.Context) (string, []string, error) {
	var forwardedFor string
	md, _ := metadata.FromIncomingContext(ctx)
	if md != nil {
		vs := md.Get(forwardedForMDKey)
		switch len(vs) {
		case 0:
		case 1:
			forwardedFor = vs[0]
		default:
			return "", []string{}, errors.New("Got several x-forwarded-for")
		}
	}
	var realIP string
	var proxyIPs []string
	ipsl := strings.Split(forwardedFor, ", ")
	realIP = ipsl[0]
	if len(ipsl) > 1 {
		proxyIPs = ipsl[1:]
	}

	return realIP, proxyIPs, nil
}

// UserAgent returns user agent from context gRPC MetaData.
func UserAgent(ctx context.Context) (string, error) {
	var userAgent string
	md, _ := metadata.FromIncomingContext(ctx)
	if md != nil {
		vs := md.Get(userAgentMDKey)
		switch len(vs) {
		case 0:
		case 1:
			userAgent = vs[0]
		default:
			return "", errors.New("Got several user agent")
		}
	}
	return userAgent, nil
}

// DeviceID returns device ID from context gRPC MetaData.
func DeviceID(ctx context.Context) (string, error) {
	var deviceID string
	md, _ := metadata.FromIncomingContext(ctx)
	if md != nil {
		vs := md.Get(deviceIDMDKey)
		switch len(vs) {
		case 0:
		case 1:
			deviceID = vs[0]
		default:
			return "", errors.New("Got several device IDs")
		}
	}
	return deviceID, nil
}

// SessionID returns session ID from context gRPC MetaData.
func SessionID(ctx context.Context) (string, error) {
	var sessionID string
	md, _ := metadata.FromIncomingContext(ctx)
	if md != nil {
		vs := md.Get(sessionIDMDKey)
		switch len(vs) {
		case 0:
		case 1:
			sessionID = vs[0]
		default:
			return "", errors.New("Got several sessions")
		}
	}
	return sessionID, nil
}

// RequestID returns request ID from context gRPC MetaData.
func RequestID(ctx context.Context) (string, error) {
	var requestID string
	md, _ := metadata.FromIncomingContext(ctx)
	if md != nil {
		vs := md.Get(requestIDMDKey)
		switch len(vs) {
		case 0:
		case 1:
			requestID = vs[0]
		default:
			return "", errors.New("Got several request IDs")
		}
	}
	return requestID, nil
}
