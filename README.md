# livy-apl
My own subset of APL, using US-ASCII names instead of weird APL characters, written in Go.

## Hints:

*   Variables start with uppercase.
*   Functions (monadic or dyadic) start with lowercase or are special symbols.
*   The only scalars are numbers.  No chars or char strings (yet).
*   All numbers are complex128.  Enter complex constants like `4+j3` or `8-j5`.
*   Abbreviations for `iota` and `rho` are `i` and `p`.
*   Index Origin is 0, not 1.  As a special case, `iota1` or `i1` generates vectors starting with 1.
*   For outer product use two dots (instead of a small circle followed by a dot) followed by the operator. Try `(i 9) ..+ i 9`
*   Dyadic operator `j` composes complex numbers from real and imaginary parts.  Try `7 8 9 j 1 2 3` and `7 8 9 ..j 1 2 3`
*   Dyadic operator `rect` (that seems misnamed, but that's what the Go library calls it!) forms complex numbers from magnitude and angle: `5 rect Pi` is very close to -5.
*   `)v` shows variables.
*   `)m` shows monadic operators.
*   `)d` shows dyadic operators.
*   You can use `;` to separate expressions, which evaluate left to right, and have the value of the last expression.
*   As in APL, all other operators bind right to left.
*   To define monadic operator `foo X` with local vars A and B: `def foo X ; A ; B { A=100; B=iota X; A + B }`
*   To define dyadic operator `X foo Y` with local vars A and B: `def X foo Y ; A ; B { A=Y*Y; B=iota X; A + B }`
*   You can also have an integer dimension: `def X rotate[D] Y { X rot[D] Y }`
*   You can also define operators with symbol names: `def X <+> Y { sqrt (X*X) + (Y*Y) } ; 3 <+> 4` results in 5.
*   History is available (use Up and Down arrows) and it is saved in `~/.livy-apl.history` for you.
*   My reference for fancy operators is the documentation for IBM APL\360.
*   Many more operators come from Go language packages `math` and `math/cmplx` and have the same names.
*   Instead of a circle operator, trig and log functions have ASCII names.
*   A dimension to an operator must be an integer.  To laminate, use `laminate` like `(i 4) laminate (i 4)` or `(i 4) laminate[0] (i 4)`.
*   User-defined functions are one line, so there is no GOTO operator.
*   There is special syntax for a conditional expression: `1000 + if 4<6 then 10 else 90 fi + 1` evaules to 1011.
*   There is special syntax for a while loop: `X=0; I=100; while I > 0 do X = X + I; I = I - 1 done ; X` evaluates to 5050.

## Missing:

You cannot save or load a workspace.  But the history mechanism can help you rebuild things.

Missing operators include
*   Radix list conversions.
*   Format.
*   Eval.
*   Represent.

## Future:

*   Some day I'd like to have lambda expressions, like in APL2.
*   Some day I'd like to have operators as arguments to higher-level functions, like in APL2.
*   Some day I'd like to have nested matrices, like in APL2.  You might find a bit of this is present already.
*   Some day I'd like to have a bridge to stuff written in Go.  You might find a bit of this is present already.
*   Some day I'd like to have chars and char strings.  You might find a bit of this is present already.

## Example:

Here's a sample transcript, building an expression to compute prime numbers.
The lines starting with 6 spaces are what the user types.
The lines like `_0 = (*livy.Mat) 8 rho`
tell you the result is put into a temporary variable `_0`
and the shape of the result is `8` (and the internal Go datatype is `livy.Mat`).
```
$ go run livy.go 
      N=8 ; V=iota1 N ; V
   _0 = (*livy.Mat) 8 rho
1  2  3  4  5  6  7  8  
      N=8 ; V=iota1 N ; V ..+ V
   _1 = (*livy.Mat) 8 8 rho
2  3   4   5   6   7   8   9   
3  4   5   6   7   8   9   10  
4  5   6   7   8   9   10  11  
5  6   7   8   9   10  11  12  
6  7   8   9   10  11  12  13  
7  8   9   10  11  12  13  14  
8  9   10  11  12  13  14  15  
9  10  11  12  13  14  15  16  
      N=8 ; V=iota1 N ; V ..* V
   _2 = (*livy.Mat) 8 8 rho
1  2   3   4   5   6   7   8   
2  4   6   8   10  12  14  16  
3  6   9   12  15  18  21  24  
4  8   12  16  20  24  28  32  
5  10  15  20  25  30  35  40  
6  12  18  24  30  36  42  48  
7  14  21  28  35  42  49  56  
8  16  24  32  40  48  56  64  
      N=8 ; V=iota1 N ; V ..mod V
   _3 = (*livy.Mat) 8 8 rho
0  1  1  1  1  1  1  1  
0  0  2  2  2  2  2  2  
0  1  0  3  3  3  3  3  
0  0  1  0  4  4  4  4  
0  1  2  1  0  5  5  5  
0  0  0  2  1  0  6  6  
0  1  1  3  2  1  0  7  
0  0  2  0  3  2  1  0  
      N=8 ; V=iota1 N ; 0 == V ..mod V
   _4 = (*livy.Mat) 8 8 rho
1  0  0  0  0  0  0  0  
1  1  0  0  0  0  0  0  
1  0  1  0  0  0  0  0  
1  1  0  1  0  0  0  0  
1  0  0  0  1  0  0  0  
1  1  1  0  0  1  0  0  
1  0  0  0  0  0  1  0  
1  1  0  1  0  0  0  1  
      N=8 ; V=iota1 N ; +/ 0 == V ..mod V
   _5 = (*livy.Mat) 8 rho
1  2  2  3  2  4  2  4  
      N=8 ; V=iota1 N ; 2 == +/ 0 == V ..mod V
   _6 = (*livy.Mat) 8 rho
0  1  1  0  1  0  1  0  
      N=8 ; V=iota1 N ; (2 == +/ 0 == V ..mod V) compress V
   _7 = (*livy.Mat) 4 rho
2  3  5  7  
      N=200 ; V=iota1 N ; (2 == +/ 0 == V ..mod V) compress V
   _8 = (*livy.Mat) 46 rho
2  3  5  7  11  13  17  19  23  29  31  37  41  43  47  53  59  61  67  71  73  79  83  89  97  101  103  107  109  113  127  131  137  139  149  151  157  163  167  173  179  181  191  193  197  199  
      def primes N ; V { V=iota1 N ; (2 == +/ 0 == V ..mod V) compress V }
   _1 = (*livy.Box) 
"def" 
      primes 100
   _2 = (*livy.Mat) 25 rho
2  3  5  7  11  13  17  19  23  29  31  37  41  43  47  53  59  61  67  71  73  79  83  89  97  
      )v
  E : 2.718281828459045 
  J : +j1 
  N : 200 
Phi : 1.618033988749895 
 Pi : 3.141592653589793 
Tau : 6.283185307179586 
  V : [200 ]{1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60 61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80 81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100 101 102 103 104 105 106 107 108 109 110 111 112 113 114 115 116 117 118 119 120 121 122 123 124 125 126 127 128 129 130 131 132 133 134 135 136 137 138 139 140 141 142 143 144 145 146 147 148 149 150 151 152 153 154 155 156 157 158 159 160 161 162 163 164 165 166 167 168 169 170 171 172 173 174 175 176 177 178 179 180 181 182 183 184 185 186 187 188 189 190 191 192 193 194 195 196 197 198 199 200 } 
      )m
+ , - abs acos acosh asin asinh atan atanh b b2s box cbrt ceil conjugate cos cosh div double down ei erf erfc erfcinv erfinv es exp exp2 expm1 fft floor gamma gi gs i i1 ifft imag image inf iota iota1 isInf isNaN j ki ks log log10 log1p log2 mi micros millis ms nanos neg not p phase pi picos primes ps real rect rho rot round round1 round2 round3 round4 round5 round6 round7 round8 round9 roundToEven s2b sgn sin sinh sqrt square tan tanh tcl ti transpose ts u unbox up y0 y1 
      )d
!= * ** + , - / < <= == > >= \ and atan compress copysign dim div drop e expand hypot isInf j jn laminate member mod or p rect remainder rho rot take transpose xor yn 
      *EOF*
```

Notice `primes` is now in the list of monadic operators printed by `)m`.

Rob Pike named his APL `ivy`.  (It's also written in Go and uses US-ASCII names for operators.)
That inspired me to write an APL, and I named it `livy`, after the Roman historian.
APL is a "classical language", like Latin, so I think if ancient Romans programmed, they would do it in APL.
