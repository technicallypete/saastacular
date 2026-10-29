package objectstore

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Config is the S3-compatible adapter's configuration: the existing scoped OBJECT_STORE_* settings plus the trusted
// signing endpoint (OBJECT_STORE_SIGNING_ENDPOINT; empty means the data endpoint).
type Config struct {
	Endpoint        string
	SigningEndpoint string
	Region          string
	Bucket          string
	ForcePathStyle  bool
	AccessKey       string
	SecretKey       string
}

// The fixed transport (G009 Plan r1 §4.3): limits, not latency promises.
const (
	dialTimeout           = 5 * time.Second
	tlsHandshakeTimeout   = 5 * time.Second
	responseHeaderTimeout = 15 * time.Second
	idleConnTimeout       = 90 * time.Second
	maxIdleConnsPerHost   = 16
)

// The metadata keys of the binding (S3 user metadata, x-amz-meta-*).
const (
	metaObject       = "object-id"
	metaOrganization = "organization-id"
	metaGroup        = "group-id"
)

// EndpointProblems checks the endpoint transport rules (G009 Plan r1 §4.7) and returns one problem per violated rule,
// naming only the variable: each endpoint is an absolute http(s) URL with a host and an optional port, without
// userinfo, path, query or fragment; with APP_ENV=production both the data and the signing endpoint are https; an http
// data endpoint with an https signing endpoint is refused everywhere. An empty signing endpoint is the data endpoint.
// config.Load applies the same rules at startup (core/internal/config); New re-enforces them for every construction.
func EndpointProblems(data, signing, appEnv string) []string {
	var problems []string
	if signing == "" {
		signing = data
	}
	ds, dok := endpointScheme(data)
	ss, sok := endpointScheme(signing)
	if !dok {
		problems = append(problems, "OBJECT_STORE_ENDPOINT must be an absolute http or https URL with a host, an optional port and nothing else")
	}
	if !sok {
		problems = append(problems, "OBJECT_STORE_SIGNING_ENDPOINT must be an absolute http or https URL with a host, an optional port and nothing else")
	}
	if dok && sok {
		if appEnv == "production" {
			if ds != "https" {
				problems = append(problems, "OBJECT_STORE_ENDPOINT must be https with APP_ENV=production")
			}
			if ss != "https" {
				problems = append(problems, "OBJECT_STORE_SIGNING_ENDPOINT must be https with APP_ENV=production")
			}
		}
		if ds == "http" && ss == "https" {
			problems = append(problems, "OBJECT_STORE_SIGNING_ENDPOINT must not be https while OBJECT_STORE_ENDPOINT is http")
		}
	}
	return problems
}

// endpointScheme returns the scheme of a well-formed endpoint.
func endpointScheme(s string) (string, bool) {
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Hostname() == "" || u.User != nil ||
		u.Path != "" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" ||
		strings.ContainsAny(s, " \t\r\n") {
		return "", false
	}
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != p {
			return "", false
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return "", false
	}
	return u.Scheme, true
}

// S3 is the S3-compatible Store: one data client and one presign client (over the signing endpoint) with the same
// scoped credential, region and addressing, a bounded transport and no automatic retries.
type S3 struct {
	data    *s3.Client
	presign *s3.PresignClient
	bucket  string
}

var _ Store = (*S3)(nil)

// New builds the S3-compatible Store for appEnv. It refuses any configuration that breaks the endpoint transport rules
// or lacks a region, bucket or credential, even when cfg was not produced by config.Load. Errors name only variables.
func New(cfg Config, appEnv string) (*S3, error) {
	problems := EndpointProblems(cfg.Endpoint, cfg.SigningEndpoint, appEnv)
	for _, r := range []struct{ name, value string }{
		{"OBJECT_STORE_REGION", cfg.Region}, {"OBJECT_STORE_BUCKET", cfg.Bucket},
		{"OBJECT_STORE_ACCESS_KEY", cfg.AccessKey}, {"OBJECT_STORE_SECRET_KEY", cfg.SecretKey},
	} {
		if strings.TrimSpace(r.value) == "" {
			problems = append(problems, r.name+" is required")
		}
	}
	if len(problems) > 0 {
		return nil, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	signing := cfg.SigningEndpoint
	if signing == "" {
		signing = cfg.Endpoint
	}
	httpClient := awshttp.NewBuildableClient().
		WithDialerOptions(func(d *net.Dialer) { d.Timeout = dialTimeout }).
		WithTransportOptions(func(t *http.Transport) {
			t.TLSHandshakeTimeout = tlsHandshakeTimeout
			t.ResponseHeaderTimeout = responseHeaderTimeout
			t.IdleConnTimeout = idleConnTimeout
			t.MaxIdleConnsPerHost = maxIdleConnsPerHost
			t.MaxIdleConns = maxIdleConnsPerHost
		})
	client := func(endpoint string) *s3.Client {
		return s3.New(s3.Options{
			BaseEndpoint:               aws.String(endpoint),
			Region:                     cfg.Region,
			UsePathStyle:               cfg.ForcePathStyle,
			Credentials:                credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
			HTTPClient:                 httpClient,
			Retryer:                    aws.NopRetryer{},
			RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
			ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
		})
	}
	return &S3{data: client(cfg.Endpoint), presign: s3.NewPresignClient(client(signing)), bucket: cfg.Bucket}, nil
}

// unsignedPayload streams an unseekable body with UNSIGNED-PAYLOAD instead of hashing it first (no buffering).
func unsignedPayload(o *s3.Options) {
	o.APIOptions = append(o.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
}

// Put streams exactly size bytes of body to key with an explicit Content-Length, the binding metadata and
// Content-Disposition: attachment. A provider refusal is ErrDenied or ErrInvalid; any other failure, including a body
// read error or cancellation, is ErrOutcomeUnknown (the object may exist).
func (c *S3) Put(ctx context.Context, key Key, body io.Reader, size int64, contentType string, b Binding) error {
	if !key.Valid() || size < 0 || !b.valid() || body == nil || contentType == "" {
		return ErrInvalid
	}
	meta := map[string]string{metaObject: b.Object, metaOrganization: b.Organization}
	if b.Group != "" {
		meta[metaGroup] = b.Group
	}
	_, err := c.data.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket), Key: aws.String(key.s), Body: body, ContentLength: aws.Int64(size),
		ContentType: aws.String(contentType), ContentDisposition: aws.String("attachment"), Metadata: meta,
	}, unsignedPayload)
	if err == nil {
		return nil
	}
	switch status(err) {
	case http.StatusForbidden, http.StatusUnauthorized:
		return ErrDenied
	case http.StatusBadRequest, http.StatusLengthRequired, http.StatusRequestEntityTooLarge:
		return ErrInvalid
	}
	return ErrOutcomeUnknown
}

// Get returns key's info (length and binding) and its body. A missing object is ErrNotFound.
func (c *S3) Get(ctx context.Context, key Key) (Info, io.ReadCloser, error) {
	if !key.Valid() {
		return Info{}, nil, ErrInvalid
	}
	out, err := c.data.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key.s)})
	if err != nil {
		return Info{}, nil, classify(err)
	}
	if out.ContentLength == nil {
		_ = out.Body.Close()
		return Info{}, nil, ErrUnavailable
	}
	return Info{Size: *out.ContentLength, Binding: binding(out.Metadata)}, out.Body, nil
}

// Delete removes key; an absent object is success.
func (c *S3) Delete(ctx context.Context, key Key) error {
	if !key.Valid() {
		return ErrInvalid
	}
	if _, err := c.data.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key.s)}); err != nil {
		if e := classify(err); e != ErrNotFound {
			return e
		}
	}
	return nil
}

// Metadata is a Head of key: its length and binding. A missing object is ErrNotFound.
func (c *S3) Metadata(ctx context.Context, key Key) (Info, error) {
	if !key.Valid() {
		return Info{}, ErrInvalid
	}
	out, err := c.data.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(c.bucket), Key: aws.String(key.s)})
	if err != nil {
		return Info{}, classify(err)
	}
	if out.ContentLength == nil {
		return Info{}, ErrUnavailable
	}
	return Info{Size: *out.ContentLength, Binding: binding(out.Metadata)}, nil
}

// SignedGetURL presigns a GET of key for ttl over the signing endpoint, offline: the response overrides are
// attachment, the stored content type and no-store, and nothing comes from the caller's request. The URL is a bearer
// secret.
func (c *S3) SignedGetURL(ctx context.Context, key Key, ttl time.Duration, contentType string) (string, error) {
	if !key.Valid() || ttl <= 0 || contentType == "" {
		return "", ErrInvalid
	}
	req, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket), Key: aws.String(key.s), ResponseContentDisposition: aws.String("attachment"),
		ResponseContentType: aws.String(contentType), ResponseCacheControl: aws.String("no-store"),
	}, s3.WithPresignExpires(ttl))
	if err != nil || req == nil || req.Method != http.MethodGet {
		return "", ErrUnavailable
	}
	return req.URL, nil
}

func binding(m map[string]string) Binding {
	get := func(k string) string {
		for mk, v := range m {
			if strings.EqualFold(mk, k) {
				return v
			}
		}
		return ""
	}
	return Binding{Object: get(metaObject), Organization: get(metaOrganization), Group: get(metaGroup)}
}

// status returns the HTTP status of a provider response error, or 0.
func status(err error) int {
	var re *awshttp.ResponseError
	if errors.As(err, &re) {
		return re.HTTPStatusCode()
	}
	return 0
}

// classify maps a read or delete failure onto the closed errors.
func classify(err error) error {
	var nsk *types.NoSuchKey
	var nf *types.NotFound
	if errors.As(err, &nsk) || errors.As(err, &nf) {
		return ErrNotFound
	}
	switch status(err) {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusForbidden, http.StatusUnauthorized:
		return ErrDenied
	}
	return ErrUnavailable
}
