type SnakeString = slice

fun SnakeString.unpackFromSlice(mutate s: slice) {
    assert (s.remainingRefsCount() <= 1) throw 5;

    // but since here we're matching the original FunC implementation, leave no checks
    val snakeRemainder = s;
    s = createEmptySlice();     // no more left to read
    return snakeRemainder
}

fun SnakeString.packToBuilder(self, mutate b: builder) {
    b.storeSlice(self)
}

@pure
fun createBits256None(): bits256
    asm "x{0000000000000000000000000000000000000000000000000000000000000000} PUSHSLICE"

@pure
fun reserveForStorage(): void
    asm "<b x{36e6b809} s, 0 8 u, 10000000 Gram, 0 1 u, <b b> ref, b> PUSHREF c5 POP";