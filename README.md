# Benchstatn

This is a fork of benchstat, which provides options to compare more than two
benchmarks at the same time.

Benchstatn computes and compares statistics about benchmarks. It allows you
to display information about one benchmark, compare to benchmarks or make a
matrix view of multiple benchmarks all compared against best case.

Usage:

    benchstatn [options] old.txt [new.txt] [more.txt ...]

Run `benchstatn -h` for the list of supported options.

Each input file should contain the concatenated output of a number of runs
of `go test -bench`. For each different benchmark listed in an input file,
benchstatn computes the mean, minimum, and maximum run time, after removing
outliers using the interquartile range rule.

If invoked on a single input file, benchstatn prints the per-benchmark
statistics for that file.

If invoked on a pair of input files, benchstatn adds to the output a column
showing the statistics from the second file and a column showing the percent
change in mean from the first to the second file. Next to the percent
change, benchstatn shows the p-value and sample sizes from a test of the two
distributions of benchmark times. Small p-values indicate that the two
distributions are significantly different. If the test indicates that there
was no significant change between the two benchmarks (defined as p > 0.05),
benchstatn displays a single ~ instead of the percent change.

If invoked on a three or more input files, benchstatn outputs just the raw 
statistics from all the files. If you want to analyze differences, use the
-showdelta options, this compares every benchmark to the best case sample.

If you want to compare against worst case, use the -deltaworst option.

The -delta-test option controls which significance test is applied: utest
(Mann-Whitney U-test), ttest (two-sample Welch t-test), or none. The default
is the U-test, sometimes also referred to as the Wilcoxon rank sum test.

If invoked on more than two input files, benchstatn prints the per-benchmark
statistics for all the files, showing one column of statistics for each
file, with no column for percent change or statistical significance.

The -html option causes benchstatn to print the results as an HTML table.

## Example

Suppose we collect benchmark results from running `go test -bench=Encode`
five times before and after a particular change.

The file old.txt contains:

    BenchmarkGobEncode   	100	  13552735 ns/op	  56.63 MB/s
    BenchmarkJSONEncode  	 50	  32395067 ns/op	  59.90 MB/s
    BenchmarkGobEncode   	100	  13553943 ns/op	  56.63 MB/s
    BenchmarkJSONEncode  	 50	  32334214 ns/op	  60.01 MB/s
    BenchmarkGobEncode   	100	  13606356 ns/op	  56.41 MB/s
    BenchmarkJSONEncode  	 50	  31992891 ns/op	  60.65 MB/s
    BenchmarkGobEncode   	100	  13683198 ns/op	  56.09 MB/s
    BenchmarkJSONEncode  	 50	  31735022 ns/op	  61.15 MB/s

The file new.txt contains:

    BenchmarkGobEncode   	 100	  11773189 ns/op	  65.19 MB/s
    BenchmarkJSONEncode  	  50	  32036529 ns/op	  60.57 MB/s
    BenchmarkGobEncode   	 100	  11942588 ns/op	  64.27 MB/s
    BenchmarkJSONEncode  	  50	  32156552 ns/op	  60.34 MB/s
    BenchmarkGobEncode   	 100	  11786159 ns/op	  65.12 MB/s
    BenchmarkJSONEncode  	  50	  31288355 ns/op	  62.02 MB/s
    BenchmarkGobEncode   	 100	  11628583 ns/op	  66.00 MB/s
    BenchmarkJSONEncode  	  50	  31559706 ns/op	  61.49 MB/s
    BenchmarkGobEncode   	 100	  11815924 ns/op	  64.96 MB/s
    BenchmarkJSONEncode  	  50	  31765634 ns/op	  61.09 MB/s

The order of the lines in the file does not matter, except that the output
lists benchmarks in order of appearance.

If run with just one input file, benchstatn summarizes that file:

    $ benchstatn old.txt
    name        time/op
    GobEncode   13.6ms ± 1%
    JSONEncode  32.1ms ± 1%

If run with two input files, benchstatn summarizes and compares:

    $ benchstatn old.txt new.txt
    name        old time/op  new time/op  delta
    GobEncode   13.6ms ± 1%  11.8ms ± 1%  -13.31% (p=0.016 n=4+5)
    JSONEncode  32.1ms ± 1%  31.8ms ± 1%     ~    (p=0.286 n=4+5)

Note that the JSONEncode result is reported as statistically insignificant
instead of a -0.93% delta.
