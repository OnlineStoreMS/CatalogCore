package storage

import "testing"

func TestPublicURLResolverRewritesHost(t *testing.T) {
	r := &PublicURLResolver{
		baseURL:   "https://osms.zfcycle.com/minio/catalogcore",
		keyPrefix: "uploads",
	}
	in := "http://192.168.3.41:9100/catalogcore/uploads/images/foo.jpg"
	want := "https://osms.zfcycle.com/minio/catalogcore/uploads/images/foo.jpg"
	if got := r.Resolve(in); got != want {
		t.Fatalf("Resolve() = %q, want %q", got, want)
	}
}
