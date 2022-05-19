package grpcurl

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

type Client struct {
	address   string
	userAgent string
	conn      *grpc.ClientConn
}

func String(s string) *string {
	return &s
}

func New(address string, userAgent *string) *Client {
	if userAgent == nil {
		userAgent = String("terraform-provider-grpc")
	}

	return &Client{
		address:   address,
		userAgent: *userAgent,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	conn, err := dial(ctx, c.address)
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

type InvokeRPCOptions struct {
	Format string
}

func dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	tlsConf, err := grpcurl.ClientTLSConfig(false, "", "", "")
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(tlsConf)

	return grpcurl.BlockingDial(ctx, "tcp", target, creds, opts...)
}

func (c Client) InvokeRPC(ctx context.Context, method string, headers []string, body string, opts InvokeRPCOptions) ([]byte, error) {
	if c.conn == nil {
		return []byte{}, fmt.Errorf("client connection is nil, run client.Connect to open a connection")
	}

	var b bytes.Buffer

	if opts.Format == "" {
		opts.Format = "json"
	}

	refCtx := metadata.NewOutgoingContext(ctx, grpcurl.MetadataFromHeaders(headers))

	refClient := grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(c.conn))

	defer refClient.Reset()

	source := grpcurl.DescriptorSourceFromServer(ctx, refClient)

	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format(opts.Format), source, strings.NewReader(body), grpcurl.FormatOptions{
		EmitJSONDefaultFields: true,
		IncludeTextSeparator:  true,
		AllowUnknownFields:    true,
	})
	if err != nil {
		return b.Bytes(), fmt.Errorf("Failed to construct request parser and formatter for %q: %w", opts.Format, err)
	}

	h := &grpcurl.DefaultEventHandler{
		Out:            &b,
		Formatter:      formatter,
		VerbosityLevel: 0,
	}

	if err := grpcurl.InvokeRPC(ctx, source, c.conn, method, headers, h, rf.Next); err != nil {
		if errStatus, ok := status.FromError(err); ok {
			h.Status = errStatus
		} else {
			return []byte{}, fmt.Errorf("Error invoking method %q: %s", method, err)
		}
	}

	if h.Status.Code() != 0 {
		var eb bytes.Buffer

		grpcurl.PrintStatus(&eb, h.Status, formatter)
		return []byte{}, fmt.Errorf(string(eb.Bytes()))
	}

	return b.Bytes(), nil
}
