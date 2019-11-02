UNIQID=$1

rm -f $UNIQID-get.out
echo "binary" > $UNIQID-get.out
echo "get $UNIQID-test-even.dat" >> $UNIQID-get.out
echo "get $UNIQID-test-odd.dat" >> $UNIQID-get.out
echo "quit" >> $UNIQID-get.out
cat $UNIQID-get.out | tftp 127.0.0.1  > /dev/null


md5sum $UNIQID-test-even.dat > $UNIQID-get-md5sum.out
md5sum $UNIQID-test-odd.dat >> $UNIQID-get-md5sum.out

rm -f $UNIQID-test-even.dat $UNIQID-test-odd.dat $UNIQID-get.out

