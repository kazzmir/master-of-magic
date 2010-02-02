(*  Example program for the Allegro library, by Shawn Hargreaves.
 *
 *  This program demonstrates how to use fixed point numbers, which
 *  are signed 32-bit integers storing the integer part in the
 *  upper 16 bits and the decimal part in the 16 lower bits. This
 *  example also uses the unusual approach of communicating with
 *  the user exclusively via the allegro_message() function.
 *)

open Allegro ;;

let () =
   allegro_init();

   (* convert integers to fixed point like this *)
   let x = itofix 10 in

   (* convert floating point to fixed point like this *)
   let y = ftofix 3.14 in

   (* fixed point variables can be assigned, added, subtracted, negated,
    * and compared just like integers, eg: 
    *)
   let z = fixadd x y in
   allegro_message(Printf.sprintf "%f + %f = %f\n" (fixtof x) (fixtof y) (fixtof z));

   (* fixed point variables can be multiplied or divided by integers or
    * floating point numbers, eg:
    *)
   let z = fixmul y (itofix 2) in
   allegro_message(Printf.sprintf "%f * 2 = %f\n" (fixtof y) (fixtof z));

   let z = fixmul x y in
   allegro_message(Printf.sprintf "%f * %f = %f\n" (fixtof x) (fixtof y) (fixtof z));

   (* fixed point trig and square root are also available, eg: *)
   let z = fixsqrt x in
   allegro_message(Printf.sprintf "fixsqrt(%f) = %f\n" (fixtof x) (fixtof z));
;;

