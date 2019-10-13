// Code generated by "esc -o ./specfs/fs.go -pkg specfs swagger.json"; DO NOT EDIT.

package specfs

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.isDir {
		return nil, fmt.Errorf(" escFile.Readdir: '%s' is not directory", f.name)
	}

	fis, ok := _escDirs[f.local]
	if !ok {
		return nil, fmt.Errorf(" escFile.Readdir: '%s' is directory, but we have no info about content of this dir, local=%s", f.name, f.local)
	}
	limit := count
	if count <= 0 || limit > len(fis) {
		limit = len(fis)
	}

	if len(fis) == 0 && count > 0 {
		return nil, io.EOF
	}

	return fis[0:limit], nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		_ = f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/swagger.json": {
		name:    "swagger.json",
		local:   "swagger.json",
		size:    18005,
		modtime: 1570908390,
		compressed: `
H4sIAAAAAAAC/+x8W3PbRpbwu3/FKc6XykyVJHo8mXy1ftl1bCujKk+ileyaB0vlOuw+JDpqdMPdDdGo
lP/71ulugAABkpLiaDPZ+MUU2Jdzv4M/PwGYCWt8XZKfPYf3TwAAZlhVWgkMypr5T96a2ROA6yNeWzkr
a3G3tX6NqxW52XOYPTt5OovPlFna2XP4Oe2V5IVTFe/lVW8Lgqp2lfUEdgmhUB56x4PyECxUzt4qSfDi
/AzsLTn4x9u35/xFsKuVJgee3K0SdHRllIF1oUQBja1BoAFlAjkUAdYqFBAKaheDMoB89MphWWJQAtbY
nESgAWZBBU0MYr7cp9tNultjQy5CUNA2FN0Rt+R8xvP2r0ykz5EgC/R0jqHg53Os1Pz2r4lSFYbCb0g1
F1qRCXNhzVKtTiKl2y8BZisKvT+nSKs8kJGVVSYA+YqEQq0bqD0taw1L6+CfdqE0wVX99Omzb+Hy/AUT
359cGeZLt5fJuCAQqDXJRMdIhO9fv4WSQmEloIc1aX10Zc5/vHzLbKs9SbDLpcq3WqMbJpivq8q6AKX1
AQq1KnQDuPCBeUQSihAqSIj7oyuzLsgR1F6ZFVTYaIuSD6mr+EmgMTZE2GxZaQoteDYU5NJZCUDfMiVS
aiT++fmUYKd/173dI4W4124vCpq4mmHtgZif+OkzAq5GBzjShJ6OlxpXw4MqpW3YAUxdlugaFpcLCrUz
gFpHmeZjwAcM5CEUGGCNHhx9rMkzlZVJq/DWOlZchHO+ZUBlW5GL1DiTfMHLyNSXUZj7yyp0WFIgt43S
z73PADODZdTH76xsBghGG8PfLCa+ieTGgaK0RGyqeJ5d/EQibO2LBP1YK0dyC6r2RjnasWHCB6Zej3cj
wm8kqSIXFPkJANMlU8/Huh6J//pTOHsVLWY0qgutBNRGfQQlEwuVJBPUsokLojY7SMJxNHVHSyAfnDKr
6TWfjlf2uOXMBorpxfQJWUujUeWlx/QpkDOoj5U8jlb7WPljBvlYmeNQ0LFvfKByNjrs80Ha341uF2nP
KW8BdBQpo5UPLNOVU7cYCPLBSSUY10RNX9haSzY+JQZRkARcoTI+JPozgoBGJiWKz8g4q3VJJkTjSyiK
k/2kR+ewmV6iApW70Bwzb2LR58MM7ZPnEEvfT8NRNscDyzSx6nrM3if7/u7/9XnSrDnylTV+pFazZ0+f
jkg2+3+Oloztn+bdvux6k7G6yE/7sA+IN/vmrseSc9YdPu/vv/y8J9uf0v/5ntk8M2XOTJkLRxhoFGFU
1t8hxOjMPPsOu/Yx8goW0qGAYGgNLPi2DrAkDLVLuvT7d8kJ6eEZ09qwwym/TDRkBbzIFDzNFDwd+vmR
t40786a8Z3vL/5bn3UiypKUyiqH289OReXgURd9Bp9+10it/nBXxmAwuNMn76/4FVbqBBYobWBcU4+0Y
EbYKbh0grNQtmewLlYSYkMT7wDowNpxcme8aeEVLrHU4SunfWmnNXlWyZJbKcEzfv4C9MH1SPhxdmehp
PdXSgkMjbQmSMztWvOxt/cYbH10Z60BxlslJBZoGSjR1zE48hb57Zp9d+4hEt/23khX9kckcMpoFiRvY
tpeXiaWn1qVUZY/tPPN50+skrL9dm7kN6UVK0M5Tpvz41nSjQv9X7OdK2wVr8q9mSJmgbDPqAO1dG+MZ
AzC2qpPVHYGePKQSCgdl0pqvAxR4S51JPrkyl7Uo4lI+aMGpDFTOCvJemdUR3yXR3YDG2ogCllrFAtEv
MttvCzIw+LrNsZRkKIlNIt2Sa/hzwhYWDVTkBJmAK/rDGP8ejDEL6fdRqJkXd7HI32cV+PexzFsQZwsd
L//DPH9Z87z2e+vjnkTtVGj2iwdW6sMNxUXXB1myt+QuaufIBN1ASWhCrFuzxVk0sVNAzsf6ULZtrUFJ
hg0cobds0wplZCrqYQAV+GPJFru0joDQR4slUItac6oYTV86Hj7WaIIKTQyU0aTv0j2b7yIMMSJe2rYA
ZlHCAjUaocxqeGNr3pUPSnQBctv5SMCvaeGtuKEAC2SEhS3L2rQtnVwl4K0tGYIFaUHbNTAORjTwsVbi
pq03p0aAEgWDoiShjjhHy9LVqbM5UUaqWyVr1N3uRLoWTHLgSJC6pejGSGDtE9om2abYhiL2NV75wLQS
1hgSDPwRvH153vs7dqMKwsghpkqJets1sdpHr8KXf4GEaXi49nbqhhQq9Kr4iVy+9YL5nOSVN3lPbJZV
jkJiVUsZjyVLxUoJ3p+EMIn+YtMLkFbUfEjc+pvxiw/xaYNHiZLDZ52AH3R/l0kPXpyfbUR2qpKTPeIe
F/iv7s7frM/rQPx3SkP+/sX93N/udt7av7Eov4tmll67+7m8rpnco/+mc/xyonodWd9zjxONeGFNtHf0
qaIY6QabgurY3VAGvC2joSrZSGazYk1AZchtNH5XZ23TUt8FX0oi8pE5L2g7jv3KcduYTDZTkIvrq2EH
cld3rVX/bemabg1BYbX0oAyjnayiI/ZScpC3lGhwRSxxQ0uxp8m4r/s329XE2oYyda4YxmSoE12yC+s3
rrqHyf+P252H2qEoZRQz1OcHupbtSQtrNaHZbvmM+k7DntPpuHX6eaeyTXWrZmOd6UQirq9Q3OAqblmp
UNSLE2HLeR7iOFa2/TinT0GZsJxzytR9wErNBmHn65GxuYeqtVkz6jU2Prn/I3C0Qic1eZ/Llb302W9l
/wXpKhU5Jd2SZtb4HGqyp0YPCKbWGhJjB4dPDK8kEbqPMo/x77BycaYgVQLMWIPrWLvoVPwOqrsx7Ie0
d6LfNC7wnkwrayu4u0Vtg/OjS1vfC20L2mFebe/uwriYBkQXl8LlLmWJwzUxfouZC4K3y7BGTj7MShka
WP6dfOODD7AswtYa/1aiAyrdma50yhexr8JKuoN5zarqA8Y0QNIAlESpXHqKqR07JG0FauWjr9htZjnS
XpEbL0h+Ji/59pvZfmv5kvEYLdl05b95+tcne4zvrCTvk3TehRR5dcK7k5wePezQEMG6iJkdP4wIp4D2
5Mr8YIcixrzCFeagIw1LyOPak8uZyj/z1SU2IAo0K+pmkGrW8KOYloRsDh2hxIXSKfmtkhuUNubajmKp
LznDbIaVie2gNTZ5dCzmyz/VPsSjYhYcTnZzc8eIzJBXGYUH+baoHI9ta2JHdrcfi3lLHjdaoFcCKJYU
Up1ho31JXAaWOAZVh33MLnMymovaBu3sFTiqHHlmbhpUE6GtAyiffWHnp3q5fhaqy3rCoOzh9pBbZ696
rOplNvn7nx906A/8/+Sx7ZzB1sm72+zZO345ecrBg287JGM5uhjBeG+uOzTygyeSHzzqgwEAGnlJJLeH
UmLOUedCV+oZGw6FulZv12LIxatUi+mHDqRT8aeNwG5R12xKai1Tc2VBZMDQSqcsatHkAdYukItDZOjb
njXjdHRlFnUW0XU8ypGsBeVN9Eklc5Ymx+SKRt2dNZpwZXLkh+BVqTTmAT9YOVtXKRj0tRDkfWwSMTDt
aM4iFgB3hERT7mqvq9oKzTM3puXXB4eBVs29Bfiy3fj4gnw5hvneAi1JKK+s+RBLax/Y/t6JAu8u3kwT
ciO8B7TjfCPlQ/2I13DSu+aUtuFwp7G1Y5lyUSF6Y49JJ1hF+hMYuvl1RGgD8mNx+0Bz/Q6MnxwhnmVP
2K/jqU4zrg9KTbt9P4fb4mL2z4Na42aGVXkwRHJT6ilI3OQ4tpdrPcwLno7wHM4AZ8PTtdKHE3Gfj+7h
7/fOP9cEZ6+6sn40h+vCfu37Rp2p4IdEOHkg2hGWrRHo8ehzgu9YyccO6u7SmPzFov2HMG8J86Owlv3C
/rLTigw51KmG20XIJEH551fmyrxPYwTPr9/P5+/Z5iuztP91XVgfrt/PrysMxfv//FiTa67f/2npcMUU
vead7y7e5EA/p1k+oMsvXSF4jb4AXIbc9ErXxIn7ruXEsVqCIgNhK/xY0477frCBNo29cwwFLBXp2DLz
wbpUsJbESXsUgfI5zL/65v9/9e3yq2dLWJCwJXmYf2/nJ1fmRczyWQOMoK1uqyor673KPbZAWuc3zCJS
5NuUIcKwJkdXZusbh2t4d/EmlljSVl4GXz1bnkCMIKXyQRnRvvHmMOaofLELaEKODWMRLcJ0lGoRJeen
TGuO+OL1UjkSQTc5aT5H5zlBzkd7Ch4WNhRpMYNzgev4OUPKUKqQ63b+6MrwmncXb772cBk1pB2mqTn2
7DYvM6mQY2ElgYywMWm2y3gVI9DE4Zw2DH7tBVYk4/Z05H1Kji8ioJ388sUVoxphhT8HEoVRfF1zBAjv
Ls7A0ZIcM/cvd6hWnVon6L9Z6nYla6PSdt9bnWY5PZDpTe79x/a02B33/Rh15SE785uI996Xmf/ArXup
u2/vZbQND9n5zpO7Y3ydDd997LehMK+d3jLI7UH7rXK7DBiNqEcGVFnWITb2yQisfK1T88kuYyyeMkgj
r0yF3q+tk12xNDXF3l28OYEXJvUMWOu6S1K6qjysanRoApGMaWNMXXFzuqcAf65sIBPSeByVVZwX8Sln
SJntxelLePa3//j2L9lY2Co1iXQTlTLBdgfdvhM9dzWYHxyvdD32u0csEgMeCFdeYcCutCisSz1fuXkD
GSvVvs3aFbfT5EcH0K7exFShexCo8OW7opTtJmMXfM3K5ng5EeG0sXcOWpVM10x3qKdmFw5Q6sfu3aFB
NVkU1veJkdxxP0anTyTqcK8gnUxdjt5T2zfrvTdc7tPgejc3OgR3Bo4jAB4raOxGCMbzFrPJ1992G7Gp
fn4bRAxa57iwdRg09t3XfvRy2KhPRz1tmJpK2WHHd45BTI4PHnwTaA/+B7Z+SVocoMSWnZjEdGI8Zjdu
E+3dx8Nmmq87Ou7TuA6mdvag2V83mEPh6LJVkYxmRjGZpQJlGoy5Uez/llA5u9BU5hi8dcttr0EZoWu5
GZtbWNk8jAqjTvAkAXaOGe2mxb8ux1t6Pf04PBnhj28NJJ7Hh+1EZ/vDDAFvKI2Nrn1vatK3lfPUNYxp
zIKgQCN1iiocBde0uQIH8lCboDSnGbF8TXLriF4rkbkl7C25FD3x95kVZZwSyr82kgr+6fWJYGHF3zHH
ndXpN0YKu4Y0+DoAvu332oo9lImVdrPS3XCpit3i+Fsk3oIX2OU87YRFPoDQK46SRKHoNiJ0tqFoKnZg
fxSXXCtvTKYEe3BofIVpzDi/iJEpgg3k38ZoOfarSBl7kSef/ycAAP//8e3GnVVGAAA=
`,
	},
}

var _escDirs = map[string][]os.FileInfo{}
