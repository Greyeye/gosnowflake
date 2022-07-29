// Example: Use AWS KMS to sign JWT and Fetch one row.
//
// using
// https://github.com/snowflakedb/gosnowflake/blob/master/cmd/select1/select1.go
//
// you must assign the matching public key to a snowflake test user.
// https://docs.snowflake.com/en/user-guide/key-pair-auth.html
//
// new env variable SNOWFLAKE_TEST_KMSARN must be provided to run the test.
// e.g.
// SNOWFLAKE_TEST_KMSARN="arn:aws:kms:us-east-1:123456789:key/mrk-123456789" selectkms1
//
// The executing user must have access to KMS' Sign permission , (e.g. iam policy of "kms:Sign")
//
// No cancel is allowed as no context is specified in the method call Query(). If you want to capture Ctrl+C to cancel
// the query, specify the context and use QueryContext() instead. See selectmany for example.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	sf "github.com/snowflakedb/gosnowflake"
)

// getDSN constructs a DSN based on the test connection parameters
func getDSN() (string, *sf.Config, error) {
	env := func(k string, failOnMissing bool) string {
		if value := os.Getenv(k); value != "" {
			return value
		}
		if failOnMissing {
			log.Fatalf("%v environment variable is not set.", k)
		}
		return ""
	}

	account := env("SNOWFLAKE_TEST_ACCOUNT", true)
	user := env("SNOWFLAKE_TEST_USER", true)
	password := env("SNOWFLAKE_TEST_PASSWORD", false)
	host := env("SNOWFLAKE_TEST_HOST", false)
	portStr := env("SNOWFLAKE_TEST_PORT", false)
	protocol := env("SNOWFLAKE_TEST_PROTOCOL", false)
	kmsarn := env("SNOWFLAKE_TEST_KMSARN", true)
	port := 443 // snowflake default port
	var err error
	if len(portStr) > 0 {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", nil, err
		}
	}
	cfg := &sf.Config{
		Account:       account,
		User:          user,
		Password:      password,
		Host:          host,
		Port:          port,
		Protocol:      protocol,
		Authenticator: sf.AuthTypeKMSJwt,
		AWSKMSKeyARN:  kmsarn,
	}
	dsn, err := sf.DSN(cfg)
	return dsn, cfg, err
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	dsn, cfg, err := getDSN()
	if err != nil {
		log.Fatalf("failed to create DSN from Config: %v, err: %v", cfg, err)
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		log.Fatalf("failed to connect. %v, err: %v", dsn, err)
	}
	defer db.Close()
	query := "SELECT 1"
	rows, err := db.Query(query) // no cancel is allowed
	if err != nil {
		log.Fatalf("failed to run a query. %v, err: %v", query, err)
	}
	defer rows.Close()
	var v int
	for rows.Next() {
		err := rows.Scan(&v)
		if err != nil {
			log.Fatalf("failed to get result. err: %v", err)
		}
		if v != 1 {
			log.Fatalf("failed to get 1. got: %v", v)
		}
	}
	if rows.Err() != nil {
		fmt.Printf("ERROR: %v\n", rows.Err())
		return
	}
	fmt.Printf("Congrats! You have successfully run %v with Snowflake DB!\n", query)
}