#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  13_matrix.ksh — Matrix operations
#
#  Topics:
#    creation · identity · get/set · add · multiply
#    transpose · determinant · pattern filling
#    practical: image transform simulation, rotation matrix
# ─────────────────────────────────────────────────────────────────────

# ── Creating matrices ─────────────────────────────────────────────────
println "=== creating matrices ==="
A = matrix(3, 3)             # 3×3 zeros
I = matrix_identity 3        # 3×3 identity
B = matrix(2, 4, 1.0)        # 2×4 filled with 1.0

println "3×3 zero matrix:"
matrix_show A
println ""
println "3×3 identity:"
matrix_show I
println ""
println "2×4 ones:"
matrix_show B
println ""

# ── Setting and getting cells ─────────────────────────────────────────
println "=== fill a matrix ==="
M = matrix(3, 3)
n = 1
for row in range(0..2) {
    for col in range(0..2) {
        matrix_set M $row $col $n
        n++
    }
}
println "3×3 counting matrix:"
matrix_show M
println ""

# Read individual cells
println "M[0][0] = $(matrix_get M 0 0)"
println "M[1][1] = $(matrix_get M 1 1)"
println "M[2][2] = $(matrix_get M 2 2)"
println ""

# ── Transpose ─────────────────────────────────────────────────────────
println "=== transpose ==="
T = matrix(2, 3)
matrix_set T 0 0 1.0 ; matrix_set T 0 1 2.0 ; matrix_set T 0 2 3.0
matrix_set T 1 0 4.0 ; matrix_set T 1 1 5.0 ; matrix_set T 1 2 6.0

println "Original (2×3):"
matrix_show T
matrix_transpose T
println "Transposed (3×2):"
matrix_show T
println ""

# ── Addition ──────────────────────────────────────────────────────────
println "=== element-wise addition ==="
X = matrix(2, 2)
Y = matrix(2, 2)
matrix_set X 0 0 1.0 ; matrix_set X 0 1 2.0
matrix_set X 1 0 3.0 ; matrix_set X 1 1 4.0
matrix_set Y 0 0 5.0 ; matrix_set Y 0 1 6.0
matrix_set Y 1 0 7.0 ; matrix_set Y 1 1 8.0

println "X:"
matrix_show X
println "Y:"
matrix_show Y
matrix_add X Y Z
println "X + Y:"
matrix_show Z
println ""

# ── Multiplication ────────────────────────────────────────────────────
println "=== matrix multiplication ==="
# 2×3 times 3×2 → 2×2
P = matrix(2, 3)
Q = matrix(3, 2)
matrix_set P 0 0 1.0 ; matrix_set P 0 1 2.0 ; matrix_set P 0 2 3.0
matrix_set P 1 0 4.0 ; matrix_set P 1 1 5.0 ; matrix_set P 1 2 6.0
matrix_set Q 0 0 7.0 ; matrix_set Q 0 1 8.0
matrix_set Q 1 0 9.0 ; matrix_set Q 1 1 10.0
matrix_set Q 2 0 11.0 ; matrix_set Q 2 1 12.0

println "P (2×3):"
matrix_show P
println "Q (3×2):"
matrix_show Q
matrix_mul P Q R
println "P × Q (2×2):"
matrix_show R
println ""

# ── Determinant ───────────────────────────────────────────────────────
println "=== determinant ==="
D2 = matrix(2, 2)
matrix_set D2 0 0 3.0 ; matrix_set D2 0 1 8.0
matrix_set D2 1 0 4.0 ; matrix_set D2 1 1 6.0
det2 = matrix_det D2
println "2×2 [[3,8],[4,6]]: det = $det2"

D3 = matrix(3, 3)
matrix_set D3 0 0 6.0 ; matrix_set D3 0 1 1.0 ; matrix_set D3 0 2 1.0
matrix_set D3 1 0 4.0 ; matrix_set D3 1 1 -2.0 ; matrix_set D3 1 2 5.0
matrix_set D3 2 0 2.0 ; matrix_set D3 2 1 8.0 ; matrix_set D3 2 2 7.0
det3 = matrix_det D3
println "3×3 det = $det3"
println ""

# ── Identity multiplication (A × I = A) ──────────────────────────────
println "=== A × I = A ==="
A2 = matrix(3, 3)
I3 = matrix_identity 3
n2 = 1
for r in range(0..2) {
    for c in range(0..2) {
        matrix_set A2 $r $c $n2
        n2++
    }
}
matrix_mul A2 I3 result
println "A2 × I3:"
matrix_show result
println ""

# ── Practical: 2D rotation matrix ─────────────────────────────────────
println "=== 2D rotation (demo) ==="
# 90-degree rotation matrix: [[0,-1],[1,0]]
R90 = matrix(2, 2)
matrix_set R90 0 0 0.0
matrix_set R90 0 1 -1.0
matrix_set R90 1 0 1.0
matrix_set R90 1 1 0.0
println "90° rotation matrix:"
matrix_show R90

# Point (3, 1) as a column vector (2×1 matrix)
vec = matrix(2, 1)
matrix_set vec 0 0 3.0
matrix_set vec 1 0 1.0
println "Point (3,1):"
matrix_show vec

matrix_mul R90 vec rotated
println "After 90° rotation:"
matrix_show rotated
println "(should be approximately (-1, 3))"
println ""

# ── Pattern: diagonal matrix from array ───────────────────────────────
println "=== diagonal matrix ==="
diag_vals = [5.0, 3.0, 7.0, 2.0]
n_size = 4
D = matrix($n_size, $n_size)
i = 0
for v in $diag_vals {
    matrix_set D $i $i $v
    i++
}
println "4×4 diagonal [5,3,7,2]:"
matrix_show D
println "det = $(matrix_det D)  (should be 5×3×7×2=210)"