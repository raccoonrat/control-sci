.PHONY: test benchmark verify verify-phase2 regression-tiny regression-full

test:
	go test ./...

benchmark:
	go test -bench=BenchmarkMediateInboundFastPath -benchmem ./tianmu/core
	go test -bench=BenchmarkNormalizerFastPath -benchmem ./tianmu/sanitize
	go test -bench=BenchmarkSessionTrackerRecordAndEvaluate -benchmem ./tianmu/core
	go test -bench=BenchmarkToolInterceptorInterceptCall -benchmem ./tianmu/toolgate

verify: test benchmark

verify-phase2: test
	go test -run 'TestRunner|TestToolInterceptorInterceptOutput' ./tianmu/detector ./tianmu/toolgate
	go test -run 'TestConfusionMatrix|TestProfiler|TestRunLiveDetectorsRegression' ./tianmu/regression
	go test -run 'TestArtifactDiff|TestExecuteArtifactDiff' ./tianmu/regression ./cmd/tianmu-regression
	go test -bench=BenchmarkBuiltinDetectors -benchmem ./tianmu/detector
	go test -bench=BenchmarkInspectAndMediateLiveDetectors -benchmem ./tianmu/detector

regression-tiny:
	go run ./cmd/tianmu-regression \
		-dataset datasets/tc260/dataset_v6/dataset_tiny.jsonl \
		-manifest datasets/tc260/dataset_v6/manifest.json \
		-out reports/tc260-v6-tiny-evidence.json

regression-full:
	go run ./cmd/tianmu-regression \
		-dataset datasets/tc260/dataset_v6/dataset.jsonl \
		-manifest datasets/tc260/dataset_v6/manifest.json \
		-out reports/tc260-v6-full-evidence.json
