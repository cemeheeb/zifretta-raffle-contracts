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