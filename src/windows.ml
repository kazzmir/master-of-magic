(*module WindowManager = sig
	class manager: object
		method paint:unit;
	end;*)

module WindowManager = struct

	class point (_x:int) (_y:int) = object(self)
		val x = _x;
		val y = _y;
		method getX = x;
		method getY = y;
	end;;

	class dimension (_width:int) (_height:int) = object(self)
		val width = _width;
		val height = _height;

		method getWidth = width;
		method getHeight = height;
	end;;

	class color (_red:int) (_green:int) (_blue:int) = object(self)
		val red = _red;
		val green = _green;
		val blue = _blue;

		method getRed = red;
		method getGreen = green;
		method getBlue = blue;
		method getRgb = (red lsl 16) + (green lsl 8) + (blue);
	end;;

	class widget = object(self)
		val mutable position = (new point 0 0)
		val mutable size = (new dimension 0 0)
		val mutable backgroundColor = (new color 0 0 0)
		(* implements event handler functions *)
		(* model *)
		(* view *)
		method paint (graphics:Graphics.AllegroGraphics.graphics) =
			graphics#fillBox position#getX position#getY size#getWidth size#getHeight backgroundColor#getRgb;
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
		(* background image *)
		(* background color *)
		(* transition in *)
		(* transition out *)
		(* mouse cursor stuff *)
	end;;

	class manager (_graphics : Graphics.AllegroGraphics.graphics) = object(self)
		inherit Graphics.eventHandler
		val graphics = _graphics
		val mutable windows:window list = [];
		val mutable currentWindow:window = ();

		(* implements event handler functions, pass to "current window" *)
		(* hash of string names to windows *)
		method addWindow (w:window) =
			windows <- w :: windows;
			currentWindow <- w;
		(* current window *)

		method paint =
			currentWindow#paint graphics;
	end;;
end;;
