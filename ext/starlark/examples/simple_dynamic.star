# Simple dynamic test generation example for PE Starlark extension
# This shows basic programmatic test creation without advanced f-string syntax

def generate_length_tests():
    """Generate a suite of length-based tests"""
    
    # Test configurations
    configs = [
        {"min": 25, "name": "very_short"},
        {"min": 50, "name": "short"}, 
        {"min": 100, "name": "medium"},
        {"min": 200, "name": "long"}
    ]
    
    tests = {}
    
    for config in configs:
        min_len = config["min"]
        name = config["name"]
        test_name = "test_min_length_" + name
        
        # Create test function dynamically  
        def length_test(response, minimum=min_len, test_name=name):
            passed = min_length(response, minimum)
            actual_len = len(response)
            
            return {
                "pass": passed,
                "reason": "Length check (" + test_name + "): " + str(actual_len) + " chars",
                "details": {
                    "minimum_required": minimum,
                    "actual_length": actual_len,
                    "test_type": test_name
                }
            }
        
        tests[test_name] = length_test
    
    return tests

def generate_content_tests():
    """Generate tests for required content elements"""
    
    required_terms = ["analysis", "example", "conclusion", "evidence"]
    tests = {}
    
    for term in required_terms:
        test_name = "test_contains_" + term
        
        def content_test(response, required_term=term):
            passed = contains(response, required_term)
            
            return {
                "pass": passed,
                "reason": "Content check: " + required_term + " " + ("FOUND" if passed else "MISSING"),
                "details": {
                    "required_term": required_term,
                    "found": passed
                }
            }
        
        tests[test_name] = content_test
    
    return tests

def test_dynamic_suite(response):
    """Run all dynamically generated tests"""
    
    # Get all test sets
    length_tests = generate_length_tests()
    content_tests = generate_content_tests()
    
    # Combine all tests
    all_tests = {}
    for name, test_func in length_tests.items():
        all_tests[name] = test_func
    for name, test_func in content_tests.items():
        all_tests[name] = test_func
    
    # Run all tests
    results = {}
    passed_count = 0
    total_count = len(all_tests)
    
    for test_name, test_func in all_tests.items():
        result = test_func(response)
        results[test_name] = result
        if result.get("pass", False):
            passed_count += 1
    
    pass_rate = float(passed_count) / float(total_count)
    
    return {
        "pass": pass_rate >= 0.5,
        "score": pass_rate,
        "reason": "Dynamic suite: " + str(passed_count) + "/" + str(total_count) + " passed",
        "details": {
            "individual_results": results,
            "pass_rate": pass_rate,
            "passed_tests": passed_count,
            "total_tests": total_count
        }
    }