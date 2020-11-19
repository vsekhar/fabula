package crc32combine

// Port of zlib's crc32_combine function.
// https://github.com/madler/zlib/blob/master/crc32.c#L372

const gf2Dim = 32

func gf2MatrixTimes(mat [gf2Dim]uint32, vec uint32) uint32 {
	var sum uint32
	i := 0
	for vec > 0 {
		if vec&1 > 0 {
			sum ^= mat[i]
		}
		vec >>= 1
		i++
	}
	return sum
}

func gf2MatrixSquare(square []uint32, mat [gf2Dim]uint32) {
	for n := 0; n < gf2Dim; n++ {
		square[n] = gf2MatrixTimes(mat, mat[n])
	}
}

// Combine takes crc1 = crc32(data1) and crc2 = crc32(data2) and returns
// crc3 = crc32(concatenate(data1, data2)). The caller must provide
// len2 = len(data2), but data1 and data2 are not needed.
//
// The polynomial poly is used. These can be obtained from the standard library
// package crc (e.g. crc.IEEE or crc.Castagnoli).
func Combine(crc1, crc2 uint32, len2 int, poly uint32) (crc3 uint32) {
	var row uint32
	var even, odd [gf2Dim]uint32 // even- and odd-power-of-two zeros operators

	// degenerate case (also disallow negative lengths)
	if len2 <= 0 {
		return crc1
	}

	// put operator for one zero bit in odd
	odd[0] = poly // was: 0xedb88320UL;          /* CRC-32 polynomial */
	row = 1
	for n := 1; n < gf2Dim; n++ {
		odd[n] = row
		row <<= 1
	}

	// put operator for two zero bits in even
	gf2MatrixSquare(even[:], odd)

	// put operator for four zero bits in odd
	gf2MatrixSquare(odd[:], even)

	// apply len2 zeros to crc1 (first square will put the operator for one zero
	// byte, eight zero bits, in even)
	for {
		// apply zeros operator for this bit of len2
		gf2MatrixSquare(even[:], odd)
		if len2&1 > 0 {
			crc1 = gf2MatrixTimes(even, crc1)
		}
		len2 >>= 1

		// if no more bits set, then done
		if len2 == 0 {
			break
		}

		// another iteration of the loop with odd and even swapped
		gf2MatrixSquare(odd[:], even)
		if len2&1 > 0 {
			crc1 = gf2MatrixTimes(odd, crc1)
		}
		len2 >>= 1

		// if no more bits set, then done
		if len2 == 0 {
			break
		}
	}

	// return combined crc
	crc1 ^= crc2
	return crc1
}
