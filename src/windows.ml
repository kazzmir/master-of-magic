module WindowManager = struct
	class widget = object(self)
		(* implements event handler functions *)
		(* model *)
		(* view *)
		(* controller *)
	end;

	class window = object(self)
		(* implements event handler functions, may pass some to "focused widget" *)
		(* list of widgets *)
		(* widget with focus *)
		(* draw: draws all widgets *)
		(* background image *)
		(* background color *)
		(* transition in *)
		(* transition out *)
		(* mouse cursor stuff *)
	end;

	class manager = object(self)
		(* implements event handler functions, pass to "current window" *)
		(* hash of string names to windows *)
		(* current window *)
	end;
end;
