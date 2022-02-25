Naming conventions for our project:
==================================

Variables: current_floor (small starting letter, divide by underscore)
Constans: MAX_FLOOR (all caps, divide by underscore)
Global variables: g_current_floor (small g, divide by underscore)
Functions to be included elsewhere: FsmOnFloorArival (regular camelcase)
Local functions : fsmOnFloorArival (small starting letter, then regular camel case)
Classes: Order (Capital letter and only one word)
Points: p_elevator_config (small p first, divide by underscore)
Enumerations/Const: en_elevator_behaviour (small en first, divide by underscore)

Includes in go
==============
Functions written with a capital letter will be included when the package that contains the function is included elsewhere
Functions written with a small starting letter will not be included
One can also define to files as being part of the same package and then they both can use each others functions no matter what
