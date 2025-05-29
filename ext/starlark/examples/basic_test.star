# Basic Starlark test example for PE
# This demonstrates simple assertion patterns

def test_min_length(response):
    """Test that response meets minimum length requirement"""
    return min_length(response, 50)

def test_contains_analysis(response):
    """Test that response contains the word 'analysis'"""
    return contains(response, "analysis")

def test_comprehensive(response):
    """Comprehensive test with scoring and detailed feedback"""
    
    score = 0.0
    reasons = []
    
    # Length check (30 points)
    if min_length(response, 100):
        score += 30
        reasons.append("Good length (100+ chars)")
    else:
        reasons.append("Too short (need 100+ chars)")
    
    # Content checks (70 points total)
    content_checks = [
        ("analysis", 25),
        ("example", 25), 
        ("conclusion", 20)
    ]
    
    for term, points in content_checks:
        if contains(response, term):
            score += points
            reasons.append(f"Contains '{term}' (+{points} points)")
        else:
            reasons.append(f"Missing '{term}' (-{points} points)")
    
    # Return detailed result
    final_score = score / 100.0
    return {
        "pass": final_score >= 0.7,
        "score": final_score,
        "reason": f"Score: {score}/100 ({final_score:.1%})",
        "details": {
            "breakdown": reasons,
            "raw_score": score,
            "max_score": 100
        }
    }

def test_word_count_range(response):
    """Test that response has appropriate word count"""
    count = word_count(response)
    
    if count < 20:
        return {
            "pass": False,
            "reason": f"Too few words: {count} (need 20-200)",
            "word_count": count
        }
    elif count > 200:
        return {
            "pass": False, 
            "reason": f"Too many words: {count} (need 20-200)",
            "word_count": count
        }
    else:
        return {
            "pass": True,
            "reason": f"Good word count: {count}",
            "word_count": count
        }

def test_quality_rubric(response):
    """Multi-dimensional quality assessment"""
    
    # Initialize scoring dimensions
    dimensions = {
        "clarity": 0,
        "completeness": 0,
        "relevance": 0
    }
    
    # Clarity assessment (0-100)
    if min_length(response, 50):
        dimensions["clarity"] += 40
    if word_count(response) >= 30:
        dimensions["clarity"] += 30  
    if not contains(response, "unclear") and not contains(response, "confusing"):
        dimensions["clarity"] += 30
    
    # Completeness assessment (0-100)
    required_elements = ["introduction", "main point", "conclusion"]
    for element in required_elements:
        if contains(response, element):
            dimensions["completeness"] += 33
    
    # Relevance assessment (0-100)
    if contains(response, "relevant") or contains(response, "related"):
        dimensions["relevance"] += 50
    if not contains(response, "off-topic") and not contains(response, "unrelated"):
        dimensions["relevance"] += 50
    
    # Calculate weighted overall score
    weights = {"clarity": 0.4, "completeness": 0.4, "relevance": 0.2}
    overall = sum(dimensions[dim] * weights[dim] for dim in dimensions) / 100
    
    return {
        "pass": overall >= 0.75,
        "score": overall,
        "reason": f"Quality assessment: {overall:.1%}",
        "dimensions": dimensions,
        "weights": weights,
        "threshold": 0.75
    }