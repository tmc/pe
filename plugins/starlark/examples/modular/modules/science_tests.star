# science_tests.star - Science test module

# Physics tests
physics_tests = [
    test(
        name = "Newton's Laws",
        input = "State Newton's three laws of motion",
        assertions = [
            icontains("rest"),
            icontains("motion"),
            icontains("force"),
            icontains("action"),
            icontains("reaction"),
            regex(r"(?i)equal|opposite"),
        ]
    ),
    
    test(
        name = "E=mc²",
        input = "Explain Einstein's equation E=mc²",
        assertions = [
            icontains("energy"),
            icontains("mass"),
            icontains("speed of light"),
            regex(r"(?i)einstein|relativity"),
        ]
    ),
    
    test(
        name = "Ohm's Law",
        input = "What is Ohm's Law and write its formula?",
        assertions = [
            regex(r"V\s*=\s*I\s*[×*]?\s*R"),
            contains("voltage"),
            contains("current"),
            contains("resistance"),
        ]
    ),
]

# Chemistry tests
chemistry_tests = [
    test(
        name = "Water molecule",
        input = "What is the chemical formula for water and describe its structure?",
        assertions = [
            contains("H2O"),
            regex(r"(?i)hydrogen"),
            regex(r"(?i)oxygen"),
            regex(r"(?i)bond|molecule"),
        ]
    ),
    
    test(
        name = "Periodic table",
        input = "Name the first 5 elements of the periodic table",
        assertions = [
            icontains("hydrogen"),
            icontains("helium"),
            icontains("lithium"),
            icontains("beryllium"),
            icontains("boron"),
        ]
    ),
    
    test(
        name = "pH scale",
        input = "Explain the pH scale and give examples of acidic and basic substances",
        assertions = [
            contains("0"),
            contains("14"),
            contains("7"),
            regex(r"(?i)acid"),
            regex(r"(?i)base|basic|alkaline"),
            regex(r"(?i)neutral"),
        ]
    ),
]

# Biology tests  
biology_tests = [
    test(
        name = "Cell types",
        input = "What are the two main types of cells and their key differences?",
        assertions = [
            icontains("prokaryotic"),
            icontains("eukaryotic"),
            regex(r"(?i)nucleus"),
            regex(r"(?i)membrane"),
        ]
    ),
    
    test(
        name = "DNA structure",
        input = "Describe the structure of DNA",
        assertions = [
            icontains("double helix"),
            regex(r"(?i)adenine|A"),
            regex(r"(?i)thymine|T"),
            regex(r"(?i)guanine|G"),
            regex(r"(?i)cytosine|C"),
            icontains("base pair"),
        ]
    ),
    
    test(
        name = "Photosynthesis",
        input = "Write the chemical equation for photosynthesis",
        assertions = [
            contains("CO2"),
            contains("H2O"),
            contains("C6H12O6"),
            contains("O2"),
            regex(r"(?i)light|sun"),
        ]
    ),
]

# Export all science tests
tests = physics_tests + chemistry_tests + biology_tests