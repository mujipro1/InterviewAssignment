#!/usr/bin/env python3
"""
Comprehensive Test Script for Transaction API
Tests all requirements from Instructions.txt
"""

import requests
import json
import time
import sys
from typing import Dict, Any

# Configuration
BASE_URL = "http://localhost:8080"
HEADERS = {"Content-Type": "application/json"}

class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    BOLD = '\033[1m'
    END = '\033[0m'

def print_header(text: str):
    print(f"\n{Colors.BOLD}{Colors.BLUE}{'='*60}")
    print(f"{text}")
    print(f"{'='*60}{Colors.END}")

def print_test(test_name: str):
    print(f"\n{Colors.BOLD}{Colors.YELLOW}🧪 TEST: {test_name}{Colors.END}")

def print_success(message: str):
    print(f"{Colors.GREEN}✅ SUCCESS: {message}{Colors.END}")

def print_error(message: str):
    print(f"{Colors.RED}❌ ERROR: {message}{Colors.END}")

def print_info(message: str):
    print(f"{Colors.BLUE}ℹ️  INFO: {message}{Colors.END}")

def make_request(method: str, url: str, headers: Dict = None, data: Dict = None) -> tuple:
    """Make HTTP request and return (response, success)"""
    try:
        if method == "GET":
            response = requests.get(url, headers=headers or {})
        elif method == "POST":
            response = requests.post(url, headers=headers or {}, json=data)
        else:
            raise ValueError(f"Unsupported method: {method}")
        
        print(f"Request: {method} {url}")
        if headers:
            print(f"Headers: {json.dumps(headers, indent=2)}")
        if data:
            print(f"Body: {json.dumps(data, indent=2)}")
        print(f"Response Status: {response.status_code}")
        
        try:
            response_json = response.json()
            print(f"Response Body: {json.dumps(response_json, indent=2)}")
        except:
            print(f"Response Body: {response.text}")
        
        return response, True
    except Exception as e:
        print_error(f"Request failed: {e}")
        return None, False

def test_health_check():
    """Test if the service is running"""
    print_test("Health Check")
    response, success = make_request("GET", f"{BASE_URL}/health")
    if success and response.status_code == 200:
        print_success("Service is running and healthy")
        return True
    else:
        print_error("Service is not responding")
        return False

def test_initial_balances():
    """Test that predefined users exist with correct initial balances"""
    print_test("Initial User Balances (Requirement: Predefined users 1,2,3)")
    
    expected_balances = {
        1: "100.00",
        2: "50.00", 
        3: "0.00"
    }
    
    all_passed = True
    for user_id, expected_balance in expected_balances.items():
        response, success = make_request("GET", f"{BASE_URL}/user/{user_id}/balance")
        if success and response.status_code == 200:
            data = response.json()
            if data.get("balance") == expected_balance and data.get("userId") == user_id:
                print_success(f"User {user_id} has correct initial balance: {expected_balance}")
            else:
                print_error(f"User {user_id} balance mismatch. Expected: {expected_balance}, Got: {data}")
                all_passed = False
        else:
            print_error(f"Failed to get balance for user {user_id}")
            all_passed = False
    
    return all_passed

def test_win_transaction():
    """Test win transaction increases balance"""
    print_test("Win Transaction (Requirement: state=win increases balance)")
    
    # Get initial balance
    response, success = make_request("GET", f"{BASE_URL}/user/1/balance")
    if not success or response.status_code != 200:
        print_error("Failed to get initial balance")
        return False
    
    initial_balance = float(response.json()["balance"])
    print_info(f"Initial balance for user 1: {initial_balance}")
    
    # Make win transaction
    headers = {**HEADERS, "Source-Type": "game"}
    transaction_data = {
        "state": "win",
        "amount": "25.50",
        "transactionId": "test_win_001"
    }
    
    response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
    if success and response.status_code == 200:
        data = response.json()
        expected_balance = initial_balance + 25.50
        if (data.get("message") == "Transaction applied successfully" and 
            float(data.get("balance")) == expected_balance and
            data.get("transactionId") == "test_win_001"):
            print_success(f"Win transaction successful. New balance: {data.get('balance')}")
            return True
        else:
            print_error(f"Win transaction response incorrect: {data}")
    else:
        print_error("Win transaction failed")
    
    return False

def test_lose_transaction():
    """Test lose transaction decreases balance"""
    print_test("Lose Transaction (Requirement: state=lose decreases balance)")
    
    # Get current balance
    response, success = make_request("GET", f"{BASE_URL}/user/1/balance")
    if not success or response.status_code != 200:
        print_error("Failed to get current balance")
        return False
    
    current_balance = float(response.json()["balance"])
    print_info(f"Current balance for user 1: {current_balance}")
    
    # Make lose transaction
    headers = {**HEADERS, "Source-Type": "server"}
    transaction_data = {
        "state": "lose",
        "amount": "15.25",
        "transactionId": "test_lose_001"
    }
    
    response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
    if success and response.status_code == 200:
        data = response.json()
        expected_balance = current_balance - 15.25
        if (data.get("message") == "Transaction applied successfully" and 
            float(data.get("balance")) == expected_balance and
            data.get("transactionId") == "test_lose_001"):
            print_success(f"Lose transaction successful. New balance: {data.get('balance')}")
            return True
        else:
            print_error(f"Lose transaction response incorrect: {data}")
    else:
        print_error("Lose transaction failed")
    
    return False

def test_insufficient_funds():
    """Test that balance cannot go negative"""
    print_test("Insufficient Funds Protection (Requirement: balance cannot go negative)")
    
    # Use user 3 who has 0.00 balance
    headers = {**HEADERS, "Source-Type": "payment"}
    transaction_data = {
        "state": "lose",
        "amount": "10.00",
        "transactionId": "test_insufficient_001"
    }
    
    response, success = make_request("POST", f"{BASE_URL}/user/3/transaction", headers, transaction_data)
    if success and response.status_code == 200:
        data = response.json()
        if (data.get("message") == "Insufficient funds" and 
            data.get("balance") == "0.00" and
            data.get("transactionId") == "test_insufficient_001"):
            print_success("Insufficient funds properly rejected")
            return True
        else:
            print_error(f"Insufficient funds response incorrect: {data}")
    else:
        print_error("Insufficient funds test failed")
    
    return False

def test_duplicate_transaction():
    """Test idempotent transactions (duplicates ignored)"""
    print_test("Duplicate Transaction Handling (Requirement: idempotent transactions)")
    
    # Get current balance
    response, success = make_request("GET", f"{BASE_URL}/user/2/balance")
    if not success or response.status_code != 200:
        print_error("Failed to get current balance")
        return False
    
    initial_balance = response.json()["balance"]
    print_info(f"Initial balance for user 2: {initial_balance}")
    
    # First transaction
    headers = {**HEADERS, "Source-Type": "game"}
    transaction_data = {
        "state": "win",
        "amount": "20.00",
        "transactionId": "duplicate_test_001"
    }
    
    response1, success1 = make_request("POST", f"{BASE_URL}/user/2/transaction", headers, transaction_data)
    if not (success1 and response1.status_code == 200):
        print_error("First transaction failed")
        return False
    
    data1 = response1.json()
    print_info(f"First transaction result: {data1}")
    
    # Duplicate transaction (same transactionId)
    response2, success2 = make_request("POST", f"{BASE_URL}/user/2/transaction", headers, transaction_data)
    if success2 and response2.status_code == 200:
        data2 = response2.json()
        if (data2.get("message") == "Duplicate transaction ignored" and 
            data2.get("balance") == data1.get("balance") and
            data2.get("transactionId") == "duplicate_test_001"):
            print_success("Duplicate transaction properly ignored")
            return True
        else:
            print_error(f"Duplicate transaction response incorrect: {data2}")
    else:
        print_error("Duplicate transaction test failed")
    
    return False

def test_source_type_validation():
    """Test Source-Type header validation"""
    print_test("Source-Type Header Validation (Requirement: game|server|payment)")
    
    # Test valid source types
    valid_sources = ["game", "server", "payment"]
    transaction_data = {
        "state": "win",
        "amount": "5.00",
        "transactionId": "source_test_001"
    }
    
    for source in valid_sources:
        headers = {**HEADERS, "Source-Type": source}
        response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
        if success and response.status_code == 200:
            print_success(f"Valid source type '{source}' accepted")
        else:
            print_error(f"Valid source type '{source}' rejected")
            return False
        
        # Update transaction ID for next test
        transaction_data["transactionId"] = f"source_test_{source}"
    
    # Test invalid source type
    headers = {**HEADERS, "Source-Type": "invalid"}
    transaction_data["transactionId"] = "source_test_invalid"
    response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
    if success and response.status_code == 400:
        print_success("Invalid source type properly rejected")
    else:
        print_error("Invalid source type not properly rejected")
        return False
    
    # Test missing source type
    response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", HEADERS, transaction_data)
    if success and response.status_code == 400:
        print_success("Missing source type properly rejected")
        return True
    else:
        print_error("Missing source type not properly rejected")
        return False

def test_amount_validation():
    """Test amount format validation"""
    print_test("Amount Format Validation (Requirement: string with up to 2 decimal places)")
    
    headers = {**HEADERS, "Source-Type": "game"}
    
    # Test valid amounts
    valid_amounts = ["10", "10.5", "10.50", "0.01", "999.99"]
    for i, amount in enumerate(valid_amounts):
        transaction_data = {
            "state": "win",
            "amount": amount,
            "transactionId": f"amount_valid_{i}"
        }
        response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
        if success and response.status_code == 200:
            print_success(f"Valid amount '{amount}' accepted")
        else:
            print_error(f"Valid amount '{amount}' rejected")
            return False
    
    # Test invalid amounts
    invalid_amounts = ["10.123", "abc", "-5.00", "10."]
    for i, amount in enumerate(invalid_amounts):
        transaction_data = {
            "state": "win",
            "amount": amount,
            "transactionId": f"amount_invalid_{i}"
        }
        response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
        if success and response.status_code in [400, 500]:  # Either validation error or processing error
            print_success(f"Invalid amount '{amount}' properly rejected")
        else:
            print_error(f"Invalid amount '{amount}' not properly rejected")
            return False
    
    return True

def test_state_validation():
    """Test state validation (win/lose only)"""
    print_test("State Validation (Requirement: win|lose only)")
    
    headers = {**HEADERS, "Source-Type": "game"}
    
    # Test valid states
    valid_states = ["win", "lose"]
    for state in valid_states:
        transaction_data = {
            "state": state,
            "amount": "5.00",
            "transactionId": f"state_valid_{state}"
        }
        response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
        if success and response.status_code == 200:
            print_success(f"Valid state '{state}' accepted")
        else:
            print_error(f"Valid state '{state}' rejected")
            return False
    
    # Test invalid state
    transaction_data = {
        "state": "invalid",
        "amount": "5.00",
        "transactionId": "state_invalid"
    }
    response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
    if success and response.status_code in [400, 500]:
        print_success("Invalid state properly rejected")
        return True
    else:
        print_error("Invalid state not properly rejected")
        return False

def test_user_id_validation():
    """Test user ID validation"""
    print_test("User ID Validation (Requirement: positive integer)")
    
    headers = {**HEADERS, "Source-Type": "game"}
    transaction_data = {
        "state": "win",
        "amount": "5.00",
        "transactionId": "userid_test"
    }
    
    # Test invalid user IDs
    invalid_user_ids = ["0", "-1", "abc", "1.5"]
    for user_id in invalid_user_ids:
        response, success = make_request("POST", f"{BASE_URL}/user/{user_id}/transaction", headers, transaction_data)
        if success and response.status_code == 400:
            print_success(f"Invalid user ID '{user_id}' properly rejected")
        else:
            print_error(f"Invalid user ID '{user_id}' not properly rejected")
            return False
    
    # Test non-existent user
    response, success = make_request("POST", f"{BASE_URL}/user/999/transaction", headers, transaction_data)
    if success and response.status_code in [404, 500]:  # User not found
        print_success("Non-existent user properly handled")
        return True
    else:
        print_error("Non-existent user not properly handled")
        return False

def test_balance_format():
    """Test that balance is returned as string with 2 decimal places"""
    print_test("Balance Format (Requirement: string with 2 decimal places)")
    
    response, success = make_request("GET", f"{BASE_URL}/user/1/balance")
    if success and response.status_code == 200:
        data = response.json()
        balance = data.get("balance")
        if isinstance(balance, str) and "." in balance and len(balance.split(".")[1]) == 2:
            print_success(f"Balance format correct: '{balance}'")
            return True
        else:
            print_error(f"Balance format incorrect: '{balance}'")
    else:
        print_error("Failed to get balance")
    
    return False

def test_json_responses():
    """Test that all responses are in JSON format"""
    print_test("JSON Response Format (Requirement: all responses in JSON)")
    
    # Test successful balance request
    response, success = make_request("GET", f"{BASE_URL}/user/1/balance")
    if success:
        try:
            response.json()
            print_success("Balance endpoint returns valid JSON")
        except:
            print_error("Balance endpoint does not return valid JSON")
            return False
    
    # Test successful transaction
    headers = {**HEADERS, "Source-Type": "game"}
    transaction_data = {
        "state": "win",
        "amount": "1.00",
        "transactionId": "json_test_001"
    }
    response, success = make_request("POST", f"{BASE_URL}/user/1/transaction", headers, transaction_data)
    if success:
        try:
            response.json()
            print_success("Transaction endpoint returns valid JSON")
        except:
            print_error("Transaction endpoint does not return valid JSON")
            return False
    
    # Test error response
    response, success = make_request("POST", f"{BASE_URL}/user/abc/transaction", headers, transaction_data)
    if success:
        try:
            response.json()
            print_success("Error responses return valid JSON")
            return True
        except:
            print_error("Error responses do not return valid JSON")
            return False
    
    return True

def run_all_tests():
    """Run all tests and return summary"""
    print_header("COMPREHENSIVE API TEST SUITE")
    print_info("Testing all requirements from Instructions.txt")
    
    tests = [
        ("Service Health Check", test_health_check),
        ("Initial User Balances", test_initial_balances),
        ("Win Transaction Processing", test_win_transaction),
        ("Lose Transaction Processing", test_lose_transaction),
        ("Insufficient Funds Protection", test_insufficient_funds),
        ("Duplicate Transaction Handling", test_duplicate_transaction),
        ("Source-Type Header Validation", test_source_type_validation),
        ("Amount Format Validation", test_amount_validation),
        ("State Validation", test_state_validation),
        ("User ID Validation", test_user_id_validation),
        ("Balance Format Validation", test_balance_format),
        ("JSON Response Format", test_json_responses),
    ]
    
    results = []
    for test_name, test_func in tests:
        try:
            result = test_func()
            results.append((test_name, result))
        except Exception as e:
            print_error(f"Test '{test_name}' crashed: {e}")
            results.append((test_name, False))
        
        time.sleep(0.5)  # Small delay between tests
    
    # Print summary
    print_header("TEST RESULTS SUMMARY")
    passed = 0
    total = len(results)
    
    for test_name, result in results:
        if result:
            print_success(f"{test_name}")
            passed += 1
        else:
            print_error(f"{test_name}")
    
    print_header(f"FINAL SCORE: {passed}/{total} TESTS PASSED")
    
    if passed == total:
        print_success("🎉 ALL TESTS PASSED! The API implementation is complete and correct.")
        print_info("   • POST /user/{userId}/transaction endpoint working")
        print_info("   • GET /user/{userId}/balance endpoint working") 
        print_info("   • Idempotent transactions (duplicates ignored)")
        print_info("   • Balance cannot go negative")
        print_info("   • Source-Type header validation (game|server|payment)")
        print_info("   • Amount format validation (string with 2 decimals)")
        print_info("   • State validation (win|lose)")
        print_info("   • User ID validation (positive integer)")
        print_info("   • JSON response format")
        print_info("   • Predefined users (1,2,3) with correct balances")
        return True
    else:
        print_error(f"❌ {total - passed} tests failed. Please check the implementation.")
        return False

if __name__ == "__main__":
    print_info("Starting comprehensive API test suite...")
    print_info(f"Testing API at: {BASE_URL}")
    print_info("Make sure the service is running with: docker-compose up -d")
    
    # Wait a moment for user to read
    time.sleep(2)
    
    success = run_all_tests()
    sys.exit(0 if success else 1)
