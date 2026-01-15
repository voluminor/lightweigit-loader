package lightweigit

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

func UserAgent(obj ProviderInterface) string {
	return fmt.Sprintf("%s %s; %s (Goland %s %s)", target.GlobalName, target.GlobalVersion, obj.Type(), runtime.GOOS, runtime.GOARCH)
}

func GetJSON(obj ProviderInterface, u string, out any) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent(obj))
	req.Header.Set("Accept", "application/json")

	resp, err := HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("%s api error: %s: %s", obj.Type(), resp.Status, strings.TrimSpace(string(b)))
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// //

func BuildURL(scheme, host, path, query string) *url.URL {
	return &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawPath:  (&url.URL{Path: path}).EscapedPath(),
		OmitHost: host == "",
		RawQuery: query,
	}
}

func AddURL(u *url.URL, addToPath, setQuery string) *url.URL {
	u.Path += addToPath
	u.RawPath = (&url.URL{Path: u.Path}).EscapedPath()
	u.RawQuery = setQuery
	return u
}

// //

func Marshal(m target.ModType, a any) []byte {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	enc.Encode(a)
	dataBuf := buf.Bytes()

	crc := crc32.ChecksumIEEE(dataBuf)
	dst := make([]byte, 4)
	binary.LittleEndian.PutUint32(dst, crc)

	buf.Reset()
	w, _ := flate.NewWriter(&buf, flate.BestCompression)
	w.Write(dataBuf)
	w.Close()

	var out bytes.Buffer
	out.WriteByte(byte(m))
	io.Copy(&out, bytes.NewReader(buf.Bytes()))
	out.Write(dst[:])

	return out.Bytes()
}

func Unmarshal(data []byte, a any) (target.ModType, error) {
	if len(data) < 5 {
		return 0, errors.New("not enough data to unmarshal")
	}

	m := target.ModType(data[0])

	if m.String() == "unknown" {
		return m, fmt.Errorf("unknown mod type")
	}

	crc := binary.LittleEndian.Uint32(data[len(data)-4:])

	r := flate.NewReader(bytes.NewReader(data[1 : len(data)-4]))
	var out bytes.Buffer
	io.Copy(&out, r)
	r.Close()

	data = out.Bytes()
	if crc32.ChecksumIEEE(data) != crc {
		return m, fmt.Errorf("invalid checksum")
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return m, dec.Decode(a)
}
