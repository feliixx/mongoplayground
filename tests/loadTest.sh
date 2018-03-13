echo "/run on same database, 100 documents:\n"
echo "POST http://localhost:8080/run?config=%5B%7B%22database%22%3A%22dbName%22%2C%22collection%22%3A%22collection%22%2C%22count%22%3A100%2C%22content%22%3A%7B%22fieldName%22%3A%7B%22type%22%3A%22string%22%2C%22minLength%22%3A5%2C%22maxLength%22%3A10%7D%7D%7D%5D&query=db.collection.find()" | vegeta attack -duration=5s | tee results.bin | vegeta report


echo "\nsaving one playground\n" 
curl http://localhost:8080/save/ -s -X "POST" -H "application/x-www-form-urlencoded" -d "config=%5B%7B%22database%22%3A%22dbName%22%2C%22collection%22%3A%22collection%22%2C%22count%22%3A100%2C%22content%22%3A%7B%22fieldName%22%3A%7B%22type%22%3A%22string%22%2C%22minLength%22%3A5%2C%22maxLength%22%3A10%7D%7D%7D%5D&query=db.collection.find()" -o tmp.txt

echo "\n/p/ on a saved playground\n"
echo "POST http://localhost:8080/$(cat tmp.txt)" | vegeta attack -duration=5s | tee results.bin | vegeta report


echo "POST http://localhost:8080/run" | awk '{ for (i=0; i<=100; i++) { print $0 "?config=%5B%7B%22database%22%3A%22" i "%22%2C%22collection%22%3A%22collection%22%2C%22count%22%3A100%2C%22content%22%3A%7B%22fieldName%22%3A%7B%22type%22%3A%22string%22%2C%22minLength%22%3A5%2C%22maxLength%22%3A10%7D%7D%7D%5D&query=db.collection.find()" } }' > targets.txt
echo "\nrun on different database" 

vegeta attack -lazy -duration=5s -targets=targets.txt | tee results.bin | vegeta report
