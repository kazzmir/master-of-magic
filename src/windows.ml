(*module WindowManager = sig
	class manager: object
		method paint:unit;
	end;*)

module WindowManager = struct

	class point (_x:int) (_y:int) = object(self)
		val x = _x;
		val y = _y;
	end;;

	class dimension (_width:int) (_height:int) = object(self)
		val width = _width;
		val height = _height;
	end;;

	class widget = object(self)
		val mutable position = (new point 0 0)
		val mutable size = (new dimension 0 0)
		(* implements event handler functions *)
		(* model *)
		(* view *)
		method paint =
			();
		(* controller *)
	end;;

	class window = object(self)
		inherit widget
		(*val mutable widgets : widget list = []
		val mutable currentlyFocusedWidget : widget;*)

		(* implements event handler functions, may pass some to "focused widget" *)
		(* list of widgets *)
		(* widget with focus *)
		(* draw: draws all widgets *)
		method paint =
			();
		(* background image *)
		(* background color *)
		(* transition in *)
		(* transition out *)
		(* mouse cursor stuff *)
	end;;

	class manager (_graphics : Graphics.graphics) = object(self)
		inherit Graphics.eventHandler
		val graphics = _graphics
		(*val mutable windows : window list = []
		val mutable currentWindow*)
		(* implements event handler functions, pass to "current window" *)
		(* hash of string names to windows *)
		(* current window *)

		method paint =
			currentWindow#paint graphics;
	end;;
end;;
