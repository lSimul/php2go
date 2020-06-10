# CLI examples (./)
- Backbone of the testing.
- Ordered by issues I encountered.

## 01.php
- Can be transpiled.
- Tests:
  - Function definitions.
  - Variable declarations.
  - Assignment chaining.
  - Variable renaming ($for -> for0).

## 02.php
- Can be transpiled.
- Tests:
  - Variable type redeclaration (int && string -> interface{}).

## 03.php
- Cannot be transpiled:
  - Trying to assign void to a variable.

## 04.php
- Can be transpiled.
- Tests:
  - Operator priority.

## 05.php
- Can be transpiled.
- Tests:
  - Variable scope definitions.

## 06.php
- Can be transpiled.
- Tests:
  - Out of scope lookup for variable.

## 07.php
- Can be transpiled.
- Tests:
  - Out of scope lookup for variable.
  - Lookup not stopping even though variable is already defined.
  - Moving nested definition to the outer block, using same instance of the variable.

## 08.php
- Can be transpiled.
- Tests:
  - Out of scope lookup for variable.
  - String incrementing.

## 09.php
- Can be transpiled.
- Tests:
  - Every variation of the for cycle.
    - break
    - with/without brackets
    - with/without before a call declaration

## 10.php
- Can be transpiled.
- Tests:
  - Simple conditions.
  - With/without brackets.
  - Condition does not require converting, already as bool.

## 11.php
- Can be transpiled.
- Tests:
  - Simple conditions.
  - Chaining if-else statements, both else if and elseif.
  - Conditions already as bool.

## 13.php
- Can be transpiled.
- Tests:
  - Function definition.
  - Function renaming to avoid reserved phrases.

## 14.php
- Can be transpiled.
- Tests:
  - while cycle
  - do ... while cycle
  - Conditions already as bool.

## 15.php
- Can be tranpiled.
- Tests:
  - String concatenation.
  - Combined with various data types.

## 16.php
- Can be transpiled.
- Tests:
  - Passing argument by reference.

## 17.php
- Can be tranpiled.
- Tests:
  - Converting to compatible types.
  - String concatenation.

## 18.php
- Can be transpiled.
- Tests:
  - Using switch.
  - Removing last break and adding fallthrogh when necessary.
  - Types are compatible.

## 19.php
- Can be transpiled.
- Tests:
  - String concatenating.
  - Escaping strings.
  - Not-escaped strings and binary operations.

## 20.php
- Can be transpiled.
- Tests:
  - Type conversion using double quotes.

## 21.php
- Can be transpiled.
- Tests:
  - Array creation.
  - Iteration.
  - Array access.
  - Array access nested in string.


## 22.php
- Can be traspiled.
- Tests:
  - arrray_push

## 23.php
- Can be transpiled.
- Tests:
  - foreach with arrays.
  - array_push
  - foreach with key => value.

## 24.php
- Can be transpiled.
- Tests:
  - Variable renaming.

## 25.php
- Can be transpiled.
- Tests:
  - Variable type re-definition.

## 26.php
- Can be transpiled.
- Tests:
  - Variable type re-definition.
  - Increment with interface{}.

## 27.php
- Can be transpiled.
- Tests:
  - Conditions without boolean values.

## 28.php
- Can be transpiled.
- Tests:
  - Function renaming.

## 29.php
- Can be transpiled.
- Tests:
  - Array editing.
  - Array pushing using empty square brackets.

## 30.php
- Can be transpiled.
- Tests:
  - foreach with arrays.
  - foreach with key => value.
  - Array defined at foreach body.

## 31.php
- Can be transpiled.
- Tests:
  - isset with array.
  - unset with array.

## 32.php
- Can be transpiled.
- Tests:
  - Cross function definition
- This did not used to work because I parsed the whole function, not just its definition.

## 33.php
- Can be transpiled.
- Tests:
  - Not boolean values in conditions.
  - while
- A lot of conditions are commented out because they are way too complicated, only the first one will be transpiled.

## 34.php
- Can be transpiled.
- Tests:
  - Variable variable count arguments.

## 36.php
- Can be transpiled.
- Cannot be build, 'b declared and not used'.
- Tests:
  - Access global variables using global.

## 37.php
- Can be transpiled.
- Tests:
  - Function case-insensitivity.


## 38.php
- Can be transpiled.
- Tests:
  - break and continue with defined values.

## 39.php
- Can be transpiled.
- Tests:
  - (int) conversion.
  - PHP_EOL macro conversion.

## 40.php
- Can be transpiled.
- Tests:
  - require
  - Function case-insensitivity.
  - And everything else defined in 27.php and 37.php.

## 41.php
- Can be transpiled.
- Note that this just tests conversion, mysql does not have to be present.
- Tests:
  - mysqli_connect
  - mysqli_select_db
  - mysql_query
  - mysqli_fetch_array with limited options.
  - Accessing fetched array.

## 42.php
- Can be transpiled.
- Tests:
  - microtime with limited options.
  - printf
  - cycles, conditions.

# Server examples (./server/)
- Combination of HTML and the PHP to form a web page.
- Does not bring anything new compared to CLI, it used to be critical couple commits ago.

## 1.php
- .php without `<?php?>`
- Can be transpiled.
- Tests:
  - Running as a server.
  - Running as a CLI.

## 2.php
- Can be transpiled.
- Tests:
  - For cycles in the webpage.

## 3.php
- Can be transpiled.
- Tests:
  - Arrays and function count()
  - Accessing GET parameters.
- Running both as CLI and server, GET params are just empty in the CLI.

## 4.php
- Can be transpiled.
- Tests:
  - GET parameters and foreach
  - Printing pair key => value in GET parameters.
- Running both as CLI and server, GET params are just empty in the CLI.

## 5.php
- Can be transpiled.
- Tests:
  - Access to other files, in this example simple cascading style sheet.
