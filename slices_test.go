package crate

import "testing"

func BenchmarkSlicesAppend(b *testing.B) {
	const num = 1000
	s := make([]string, 0, num)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for x := 0; x < num; x++ {
			s = append(s, "x")
		}

		s = s[:0]
	}
}

func BenchmarkSlicesSet(b *testing.B) {
	const num = 1000
	s := make([]string, num)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for x := 0; x < num; x++ {
			s[x] = "x"
		}
	}
}
