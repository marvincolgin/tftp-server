rm -f test-even.dat
rm -f test-odd.dat

dd if=/dev/random of=./test-even.dat bs=512 count=1000 2> /dev/null
dd if=/dev/random of=./test-odd.dat bs=511 count=1000 2> /dev/null

rm -f put.out
echo "binary" >> put.out
echo "put test-even.dat" >> put.out
echo "put test-odd.dat" >> put.out
echo "quit" >> put.out
cat put.out | tftp 127.0.0.1 > /dev/null

md5sum test-even.dat
md5sum test-odd.dat

rm -f test-even.dat test-odd.dat put.out
