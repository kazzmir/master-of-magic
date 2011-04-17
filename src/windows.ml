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

	type whichWindow = 
		INTROTITLE | 
		INTROCREDITS |
		MAINMENU;;

	type whichEvent =
		PAINT_EVENT 

		(* time, button, x, y *)
		| TIMER_EVENT of float * Graphics.mouseButton * float * float 

		(* time, button, x, y *)
		| MOUSE_UP_EVENT of float * Graphics.mouseButton * float * float

		(* time, button, x, y *)
		| MOUSE_DOWN_EVENT of float * Graphics.mouseButton * float * float

		(* key code *)
		| KEY_UP_EVENT of Graphics.key 

		(* key code *)
		| KEY_DOWN_EVENT of Graphics.key 
	
		(* time, button, x, y *)
		| MOUSE_CLICK_EVENT of float * Graphics.mouseButton * float * float 

		(* time, x, y *)
		| MOUSE_HOVER_EVENT of float * float * float 

		(* time, x, y *)
		| MOUSE_MOVE_EVENT of float * float * float 

		(* key code *)
		| KEYPRESS_EVENT of int 
		;;

	class widget = object(self)
		val mutable position = (new point 0 0)
		val mutable size = (new dimension 0 0)
		val mutable backgroundColor = (new color 255 255 0)
		(* implements event handler functions *)
		(* model *)
		(* view *)
		method paint (graphics:Graphics.AllegroGraphics.graphics) =
			graphics#fillBox position#getX position#getY size#getWidth size#getHeight backgroundColor#getRgb;

		(* controller *)
		method receiveEvent (m:manager) (e:whichEvent) =
			match e with
			| PAINT_EVENT -> self#paint m#getGraphics;
			| _ -> Printf.printf "What was that?\n";
	end 
    (* using `and' here makes the types mutually recursive *)
    and window = object(self)
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
	end 
    and manager (_graphics : Graphics.AllegroGraphics.graphics) = object(self)
		inherit Graphics.eventHandler

		val graphics = _graphics
		val mutable windows = Hashtbl.create 10
		val mutable currentWindow:(window option) = None

		initializer 
			self#addWindow INTROTITLE (new window);
			self#paint;

		method sendEvent (event:whichEvent) = 
			match currentWindow with
			| None -> raise (Failure "No window set")
			| Some window -> window#receiveEvent (self :> manager) event

		method mouse_down time button x y =
			self#sendEvent (MOUSE_DOWN_EVENT (time, button, x, y))

		method mouse_up time button x y = 
			self#sendEvent (MOUSE_UP_EVENT (time, button, x, y))

		method key_down a = 
			self#sendEvent (KEY_DOWN_EVENT a)

		method key_up a = 
			self#sendEvent (KEY_UP_EVENT a)

		method mouse_click time button x y = 
			self#sendEvent (MOUSE_CLICK_EVENT (time, button, x, y))

		method mouse_hover time x y = 
			self#sendEvent (MOUSE_HOVER_EVENT (time, x, y))

		method mouse_move time x y = 
			self#sendEvent (MOUSE_MOVE_EVENT (time, x, y))

		method keypress = 
			self#sendEvent (KEYPRESS_EVENT 0)

		(* implements event handler functions, pass to "current window" *)
		(* hash of string names to windows *)
		method addWindow (wh:whichWindow) (w:window) =
			Hashtbl.add windows wh w;
			currentWindow <- Some w;

		method paint =
			self#sendEvent PAINT_EVENT;

		method getGraphics =
			graphics;

		end;;
end;;
