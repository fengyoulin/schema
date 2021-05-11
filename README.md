# Schema #

For dynamic type creation, and data serialization in binary mode. Usually faster and smaller than JSON.

This is a benchmark compare with JSON and Gob:
```
$ go test -bench=.
goos: windows
goarch: amd64
pkg: github.com/fengyoulin/schema
cpu: Intel(R) Core(TM) i7-8650U CPU @ 1.90GHz
BenchmarkJsonEncode-8             397910              2911 ns/op
BenchmarkJsonDecode-8             182155              6803 ns/op
BenchmarkGobEncode-8              127317              9187 ns/op
BenchmarkGobDecode-8               35047             34652 ns/op
BenchmarkDecoder_Decode-8         633194              1766 ns/op
BenchmarkEncoder_Encode-8         924028              1305 ns/op
PASS
ok      github.com/fengyoulin/schema    7.735s
```
