UNIQID=$1

rm -f get-$UNIQID.out
echo "binary" > get-$UNIQID.out
echo "get test-even-$UNIQID.dat" >> get-$UNIQID.out
echo "get test-odd-$UNIQID.dat" >> get-$UNIQID.out
echo "quit" >> get-$UNIQID.out
cat get-$UNIQID.out | tftp 127.0.0.1  > /dev/null


md5sum test-even-$UNIQID.dat > get-md5sum-$UNIQID.out
md5sum test-odd-$UNIQID.dat >> get-md5sum-$UNIQID.out

rm -f test-even-$UNIQID.dat test-odd-$UNIQID.dat get-$UNIQID.out

