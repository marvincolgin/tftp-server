rm -f get.out
echo "binary" >> get.out
echo "get test-even.dat" >> get.out
echo "get test-odd.dat" >> get.out
echo "quit" >> get.out
cat get.out | tftp 127.0.0.1  > /dev/null


md5sum test-even.dat
md5sum test-odd.dat

rm -f test-even.dat test-odd.dat get.out

