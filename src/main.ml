let init graphics =
  graphics#initialize Graphics.Windowed 640 480
;;

class handler = object(self)
  inherit Graphics.eventHandler
  method mouse_down time button x y =
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse down %f %s x %f y %f\n%!" time name x y
  method mouse_up time button x y =
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse up %f %s x %f y %f\n%!" time name x y
  method key_down key = 
    Printf.printf "Key down %s\n%!" (Graphics.key_name key)
  method key_up key = 
    Printf.printf "Key up %s\n%!" (Graphics.key_name key)
  method mouse_click time button x y =
    let name = (Graphics.mouse_button_name button) in
    Printf.printf "Mouse click %f %s x %f y %f\n%!" time name x y
  method mouse_hover time x y = Printf.printf "Mouse hover %f x %f y %f\n%!" time x y
  method mouse_move time x y = Printf.printf "Mouse move %f x %f y %f\n%!" time x y
  method keypress = ignore()
end;;

let create_event_handler graphics = 
  (new handler)
;;

let main () : unit = 
  let graphics = (new Graphics.AllegroGraphics.graphics) in
  (*
  graphics#addTimer 100.0 (fun time -> Printf.printf "Timer called %f\n%!" time);
  *)
  (*let event = create_event_handler () in *)
  let windowManager = (new Windows.WindowManager.manager graphics) in
  init graphics;
  (*graphics#event_loop event;*)
  graphics#event_loop (windowManager :> Graphics.eventHandler) 
;;

(* main (); *)

class imageLibrarian = object(self)
	val water = Allegro.create_bitmap 16 16;

	initializer
		Allegro.clear_to_color ~bmp:water ~color:(Allegro.color_index 1);

	method getMapTileBitmap (plane:Gamedata.planeType) (mapTile:Gamedata.mapTile) (mapTileN:Gamedata.mapTile) (mapTileE:Gamedata.mapTile) (mapTileW:Gamedata.mapTile) (mapTileS:Gamedata.mapTile) (tick:int) = 
		water;
end;;

class mapWidget (_gameState:Gamedata.gameState) (_librarian:imageLibrarian) = object(self)
	val gameState = _gameState;
	val librarian = _librarian;
	method draw destination plane x y tick =
		Gamedata.loop 0 10 (fun j ->
			Gamedata.loop 0 10 (fun i ->
				let map = gameState#getMap plane in
				let mapTile = map#getNormalized (x+i) (y+j) in
				let mapTileN = map#getNormalized (x+i) (y+j-1) in
				let mapTileE = map#getNormalized (x+i+1) (y+j) in
				let mapTileW = map#getNormalized (x+i-1) (y+j) in
				let mapTileS = map#getNormalized (x+i) (y+j+1) in
				let mapTileBitmap = librarian#getMapTileBitmap plane mapTile mapTileN mapTileE mapTileW mapTileS tick in
				Allegro.blit ~src:mapTileBitmap ~dest:destination ~src_x:0 ~src_y:0 ~dest_x:(i*16) ~dest_y:(j*16) ~width:16 ~height:16;
			)
		);
end;;


let main2 () : unit =
	Allegro.allegro_init ();
	Allegro.set_color_depth 8;
	Allegro.install_timer ();
	Allegro.install_keyboard();
	ignore(Allegro.install_mouse());
	Allegro.set_gfx_mode Allegro.GFX_AUTODETECT_WINDOWED 320 240 0 0;
	let gamestate1 = new Gamedata.gameState in
	let pal1 = Allegro.get_desktop_palette() in
	let scr = Allegro.get_screen() in 
	let background1 = Allegro.load_bitmap "/home/drafkind/Downloads/scene_org_img.pcx" pal1 in
	(*Allegro.blit ~src:background1 ~dest:scr ~src_x:0 ~src_y:0 ~dest_x:0 ~dest_y:0 ~width:320 ~height:200;*)
	let librarian1 = (new imageLibrarian) in
	let mapWidget1 = (new mapWidget gamestate1 librarian1) in
	(gamestate1#terraform);
	mapWidget1#draw scr Gamedata.LIGHT 0 0 0;
	ignore(Allegro.readkey());;

main2 ();
