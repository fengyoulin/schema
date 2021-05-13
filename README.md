# Schema #

For dynamic type creation, and data serialization in binary mode. Usually faster and smaller than JSON.

This is a benchmark compare with JSON and Gob:
```
$ go test -bench=.
goos: windows
goarch: amd64
pkg: github.com/fengyoulin/schema
cpu: Intel(R) Core(TM) i7-8650U CPU @ 1.90GHz
BenchmarkJsonEncode-8             356102              2912 ns/op
BenchmarkJsonDecode-8             169387              6779 ns/op
BenchmarkGobEncode-8              134079              9056 ns/op
BenchmarkGobDecode-8               34712             34470 ns/op
BenchmarkDecoder_Decode-8         750661              1567 ns/op
BenchmarkEncoder_Encode-8        1000000              1100 ns/op
PASS
ok      github.com/fengyoulin/schema    7.507s
```
