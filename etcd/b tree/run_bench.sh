#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

RAW_FILE="benchmark_count3.txt"
MEDIAN_FILE="benchmark_median.txt"

GOCACHE="${GOCACHE:-/tmp/go-build}" go test -run=^$ -bench=. -benchmem -count=3 | tee "${RAW_FILE}"

awk '
/^Benchmark/ {
	k=$1
	c[k]++
	ns[k","c[k]]=$3
	bb[k","c[k]]=$5
	aa[k","c[k]]=$7
}
END {
	for (k in c) {
		delete t
		n=c[k]
		for (i=1; i<=n; i++) {
			t[i]=ns[k","i]
		}
		for (i=1; i<=n; i++) {
			for (j=i+1; j<=n; j++) {
				if (t[i] > t[j]) {
					tmp=t[i]
					t[i]=t[j]
					t[j]=tmp
				}
			}
		}

		m=t[int((n+1)/2)]
		bsum=0
		asum=0
		for (i=1; i<=n; i++) {
			bsum+=bb[k","i]
			asum+=aa[k","i]
		}
		printf "%s %.2f %.0f %.0f\n", k, m, bsum/n, asum/n
	}
}
' "${RAW_FILE}" | sort > "${MEDIAN_FILE}"

echo "Raw benchmark output  : ${RAW_FILE}"
echo "Median summary output : ${MEDIAN_FILE}"
