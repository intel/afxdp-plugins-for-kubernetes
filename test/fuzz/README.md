# Fuzz Tests

There are two fuzzing packages used to conduct the four fuzz tests which are as follows:

| Component | Function/Package Under-Test | Fuzzing Package |
| :---: | :---: | :---: |
| CNI | Network Config | go-fuzz |
| Device Plugin | GetConfig | go-fuzz |
| Device Plugin | UDS | go-fuzz |
| Device Plugin | AF-XDP | google/gofuzz |

Note: the following information is regarding the go-fuzz testes. As AF-XDP fuzz test uses google/gofuzz package a different procedure applies, please see [AF-XDP Fuzz Test](#af-xdp-fuzz-test)

To start fuzz testing, proceed to the function/package you wish to test and  run `./fuzz.sh`. `Ctrl + C` will stop the test from running.



The `fuzz.sh` script will:

 - Install [go-fuzz](https://github.com/dvyukov/go-fuzz) if necessary.
 - Build a test program that is capable of testing.
 - Execute the test programs against the function under-test.
 - Remove any remaining network namespace, logfiles and UDS sockets after the tests.

**CAUTION:** Fuzzing will result in the CNI placing a lot of randomly named log files under `/var/log/afxdp-k8s-plugins/`. These will need to be manually cleaned. The CNI has input validation that should ensure log files cannot be generated anywhere outside of this directory. Nonetheless caution is advised, and fuzzing should not be performed on a production system.

## Files and directories

This CNI fuzz test explanation will also apply to the additional tests using the go-fuzz package.

- `fuzz.sh` a script to conveniently perform all the fuzzing steps.
- `cni.go` contains a small amount of Go code capable of calling our CNI functions under test. Returns values based on outcome.
- `cni-fuzz.zip` is a go-fuzz archive created during the building of the test program.
- `outputAdd` and `outputDel` are the working directories for the tests. This is where go-fuzz dumps the output of each of the functions under test. Within each of these directories are several files and subdirectories. Of note is a `corpus/` directory that initially contains examples of good configs that would ordinarily be passed to our CNI functions. Go-fuzz learns from this and continually alters the configs in an attempt to break the functions under test. During fuzzing many new files will be added. These files contain input data that Go-fuzz determined as interesting. Also of note is the `crashers/` directory that will contain the details of any scenario that resulted in the crashing of the code under test.

## Sample output
Below is a sample output. As both CNI functions are tested in parallel it means every second line of the output is CmdAdd and CmdDel. Over time you'll see the reported stats of both functions drift apart significantly, with CmdDel doing far more executions, presumably as the more complex CmdAdd takes longer to execute and test.

Output columns as described in the go-fuzz [documentation](https://github.com/dvyukov/go-fuzz#usage):
 
 - **workers** means number of tests running in parallel (set with -procs flag).
 - **corpus** is current the number of interesting inputs the fuzzer has discovered, time in brackets says when the last interesting input was discovered.
 - **crashers** is number of discovered bugs (check out workdir/crashers dir).
 - **restarts** is the rate with which the fuzzer restarts test processes. The rate should be close to 1/10000 (which is the planned restart rate); if it is considerably higher than 1/10000, consider fixing already discovered bugs which lead to frequent restarts.
 - **execs** is total number of test executions, and the number in brackets is the average speed of test executions.
 - **cover** is number of bits set in a hashed coverage bitmap, if this number grows fuzzer uncovers new lines of code; size of the bitmap is 64K; ideally cover value should be less than ~5000, otherwise fuzzer can miss new interesting inputs due to hash collisions.
 - **uptime** is uptime of the process. This same information is also served via http (see the `-http` flag).

```
2021/10/11 15:28:23 workers: 88, corpus: 744 (0s ago), crashers: 0, restarts: 1/9214, execs: 14946610 (34597/sec), cover: 1808, uptime: 7m12s
2021/10/11 15:28:23 workers: 88, corpus: 791 (1m2s ago), crashers: 0, restarts: 1/9911, execs: 132825465 (307452/sec), cover: 1484, uptime: 7m12s
2021/10/11 15:28:26 workers: 88, corpus: 744 (3s ago), crashers: 0, restarts: 1/9238, execs: 15039689 (34573/sec), cover: 1808, uptime: 7m15s
2021/10/11 15:28:26 workers: 88, corpus: 791 (1m5s ago), crashers: 0, restarts: 1/9912, execs: 133809883 (307595/sec), cover: 1484, uptime: 7m15s
2021/10/11 15:28:29 workers: 88, corpus: 748 (0s ago), crashers: 0, restarts: 1/9221, execs: 15133042 (34549/sec), cover: 1808, uptime: 7m18s
2021/10/11 15:28:29 workers: 88, corpus: 791 (1m8s ago), crashers: 0, restarts: 1/9909, execs: 134795648 (307739/sec), cover: 1484, uptime: 7m18s
2021/10/11 15:28:32 workers: 88, corpus: 748 (3s ago), crashers: 0, restarts: 1/9212, execs: 15228726 (34531/sec), cover: 1808, uptime: 7m21s
2021/10/11 15:28:32 workers: 88, corpus: 791 (1m11s ago), crashers: 0, restarts: 1/9909, execs: 135770882 (307857/sec), cover: 1484, uptime: 7m21s
2021/10/11 15:28:35 workers: 88, corpus: 749 (1s ago), crashers: 0, restarts: 1/9192, execs: 15324387 (34513/sec), cover: 1808, uptime: 7m24s
2021/10/11 15:28:35 workers: 88, corpus: 791 (1m14s ago), crashers: 0, restarts: 1/9912, execs: 136740764 (307961/sec), cover: 1484, uptime: 7m24s
2021/10/11 15:28:38 workers: 88, corpus: 749 (4s ago), crashers: 0, restarts: 1/9193, execs: 15417980 (34491/sec), cover: 1808, uptime: 7m27s
2021/10/11 15:28:38 workers: 88, corpus: 791 (1m17s ago), crashers: 0, restarts: 1/9914, execs: 137729570 (308106/sec), cover: 1484, uptime: 7m27s
2021/10/11 15:28:41 workers: 88, corpus: 749 (7s ago), crashers: 0, restarts: 1/9227, execs: 15510676 (34467/sec), cover: 1808, uptime: 7m30s
2021/10/11 15:28:41 workers: 88, corpus: 791 (1m20s ago), crashers: 0, restarts: 1/9914, execs: 138701785 (308213/sec), cover: 1484, uptime: 7m30s
2021/10/11 15:28:44 workers: 88, corpus: 749 (10s ago), crashers: 0, restarts: 1/9244, execs: 15604844 (34447/sec), cover: 1808, uptime: 7m33s
2021/10/11 15:28:44 workers: 88, corpus: 791 (1m23s ago), crashers: 0, restarts: 1/9913, execs: 139659230 (308285/sec), cover: 1484, uptime: 7m33s
2021/10/11 15:28:47 workers: 88, corpus: 749 (13s ago), crashers: 0, restarts: 1/9245, execs: 15699400 (34427/sec), cover: 1808, uptime: 7m36s
2021/10/11 15:28:47 workers: 88, corpus: 791 (1m26s ago), crashers: 0, restarts: 1/9915, execs: 140609538 (308341/sec), cover: 1484, uptime: 7m36s
2021/10/11 15:28:50 workers: 88, corpus: 749 (16s ago), crashers: 0, restarts: 1/9273, execs: 15793560 (34407/sec), cover: 1808, uptime: 7m39s
2021/10/11 15:28:50 workers: 88, corpus: 791 (1m29s ago), crashers: 0, restarts: 1/9916, execs: 141591425 (308465/sec), cover: 1484, uptime: 7m39s
2021/10/11 15:28:53 workers: 88, corpus: 750 (0s ago), crashers: 0, restarts: 1/9269, execs: 15887390 (34387/sec), cover: 1808, uptime: 7m42s
2021/10/11 15:28:53 workers: 88, corpus: 791 (1m32s ago), crashers: 0, restarts: 1/9916, execs: 142589719 (308622/sec), cover: 1484, uptime: 7m42s
2021/10/11 15:28:56 workers: 88, corpus: 753 (0s ago), crashers: 0, restarts: 1/9255, execs: 15983576 (34372/sec), cover: 1808, uptime: 7m45s
2021/10/11 15:28:56 workers: 88, corpus: 791 (1m35s ago), crashers: 0, restarts: 1/9917, execs: 143568564 (308736/sec), cover: 1484, uptime: 7m45s
2021/10/11 15:28:59 workers: 88, corpus: 755 (0s ago), crashers: 0, restarts: 1/9225, execs: 16080242 (34358/sec), cover: 1808, uptime: 7m48s
2021/10/11 15:28:59 workers: 88, corpus: 791 (1m38s ago), crashers: 0, restarts: 1/9915, execs: 144551439 (308858/sec), cover: 1484, uptime: 7m48s
2021/10/11 15:29:02 workers: 88, corpus: 757 (0s ago), crashers: 0, restarts: 1/9216, execs: 16174110 (34339/sec), cover: 1808, uptime: 7m51s
2021/10/11 15:29:02 workers: 88, corpus: 791 (1m41s ago), crashers: 0, restarts: 1/9916, execs: 145543485 (308997/sec), cover: 1484, uptime: 7m51s
2021/10/11 15:29:05 workers: 88, corpus: 758 (0s ago), crashers: 0, restarts: 1/9239, execs: 16269971 (34324/sec), cover: 1808, uptime: 7m54s
2021/10/11 15:29:05 workers: 88, corpus: 791 (1m44s ago), crashers: 0, restarts: 1/9918, execs: 146545298 (309154/sec), cover: 1484, uptime: 7m54s
2021/10/11 15:29:08 workers: 88, corpus: 762 (0s ago), crashers: 0, restarts: 1/9266, execs: 16363768 (34304/sec), cover: 1808, uptime: 7m57s
2021/10/11 15:29:08 workers: 88, corpus: 791 (1m47s ago), crashers: 0, restarts: 1/9919, execs: 147513017 (309239/sec), cover: 1484, uptime: 7m57s
2021/10/11 15:29:11 workers: 88, corpus: 764 (1s ago), crashers: 0, restarts: 1/9257, execs: 16459960 (34290/sec), cover: 1808, uptime: 8m0s
2021/10/11 15:29:11 workers: 88, corpus: 791 (1m50s ago), crashers: 0, restarts: 1/9920, execs: 148504479 (309371/sec), cover: 1484, uptime: 8m0s
2021/10/11 15:29:14 workers: 88, corpus: 765 (2s ago), crashers: 0, restarts: 1/9290, execs: 16555058 (34274/sec), cover: 1808, uptime: 8m3s
2021/10/11 15:29:14 workers: 88, corpus: 791 (1m53s ago), crashers: 0, restarts: 1/9919, execs: 149478356 (309466/sec), cover: 1484, uptime: 8m3s
2021/10/11 15:29:17 workers: 88, corpus: 765 (5s ago), crashers: 0, restarts: 1/9316, execs: 16648823 (34256/sec), cover: 1808, uptime: 8m6s
2021/10/11 15:29:17 workers: 88, corpus: 791 (1m56s ago), crashers: 0, restarts: 1/9920, execs: 150431024 (309516/sec), cover: 1484, uptime: 8m6s
2021/10/11 15:29:20 workers: 88, corpus: 767 (1s ago), crashers: 0, restarts: 1/9282, execs: 16745489 (34243/sec), cover: 1808, uptime: 8m9s
2021/10/11 15:29:20 workers: 88, corpus: 791 (1m59s ago), crashers: 0, restarts: 1/9922, execs: 151414017 (309627/sec), cover: 1484, uptime: 8m9s
2021/10/11 15:29:23 workers: 88, corpus: 768 (1s ago), crashers: 0, restarts: 1/9262, execs: 16839048 (34225/sec), cover: 1808, uptime: 8m12s
2021/10/11 15:29:23 workers: 88, corpus: 791 (2m2s ago), crashers: 0, restarts: 1/9922, execs: 152355447 (309653/sec), cover: 1484, uptime: 8m12s
2021/10/11 15:29:26 workers: 88, corpus: 769 (1s ago), crashers: 0, restarts: 1/9247, execs: 16932634 (34206/sec), cover: 1808, uptime: 8m15s
2021/10/11 15:29:26 workers: 88, corpus: 791 (2m5s ago), crashers: 0, restarts: 1/9926, execs: 153329901 (309745/sec), cover: 1484, uptime: 8m15s
```

## AF-XDP Fuzz Test

For AFXDP fuzz testing, [google/goFuzz](https://github.com/google/gofuzz) package is utilised.

To start the AFXDP fuzz test:
- CNI and Device Plugin binaries must be created, from the root of the directory run `make build`.
- Navigate to the [/deviceplugin/afxdp](./deviceplugin/afxdp) directory, open `config.json` file and set `afxdpFuzz` field as `true`, see example below:
```
{
	"logLevel": "debug",
	"mode": "primary",
	"udsFuzz": true,
	"pools" : [
		{
			"name" : "fuzz",
			"drivers" : ["i40e"]
		}
	]
}
```
- Run `./fuzz.sh`. `Ctrl + C` will stop the test from running.


The `fuzz.sh` script will:

- Run both the CNI and Device Plugin.
- Deploy test pod `afxdp-fuzz-pod`.
- Execute the fuzzHandler in `internal/uds/uds_fuzz.go`.
- The fuzzHandler will call the imported google/gofuzz package.
- Execute generated fuzzed data to the function under-test in the AF-XDP application.
