type terrainClassification =
	WATER |
	PLAINS |
	DESERT |
	SWAMP |
	HILL |
	MOUNTAIN |
	FOREST |
	ICE;;

type terrainBonus =
	COAL |
	GOLD |
	SILVER |
	GEM |
	GAME |
	FISH;;

type terrainStructure =
	CITY |
	CAVE |
	TEMPLE |
	REDNODE |
	GREENNODE |
	BLUENODE;;

type planeType =
	LIGHT |
	DARK;;

class place = object(self)
end;;

class city = object(self)
	inherit place
end;;

class mananode = object(self)
	inherit place
end;;

class lair = object(self)
	inherit place
end;;

class armyunit = object(self)
end;;

class stack = object(self)
end;;

class player = object(self)
end;;

class mapTile (_terrain:terrainClassification) (_bonus:terrainBonus option) = object(self)
	val mutable terrain = _terrain;
	val mutable bonus = _bonus;

	method getTerrain =
		terrain;

	method getBonus =
		bonus;
end;;

class map (_width:int) (_height:int) (_plane:planeType) = object(self)
	val mutable width: int = _width;
	val mutable height: int = _height;
	val mutable plane: planeType = _plane;
	val mutable data: mapTile array = Array.make (_width*_height) (new mapTile WATER None);

	method get x y =
		data.(x + (y*width));

	method getNormalized x y =
		if (y < 0 || y >= height) then begin
			(new mapTile ICE None);
		end else begin
			self#get ((x+width) mod width) y;
		end
	method put x y m =
		data.(x + (y*width)) <- m;

end;;

let rec loop start stop what =
	if (start < stop) then begin
		what start;
		loop (start+1) stop what
	end;

class gameState = object(self)
	val mutable lightMap : map = (new map 200 100 LIGHT);
	val mutable darkMap : map = (new map 200 100 DARK);

	method terraform =
		loop 0 100 (fun y -> 
			loop 0 200 (fun x ->
				lightMap#put x y (new mapTile WATER None);
				darkMap#put x y (new mapTile WATER None);
				)
		);

	method getMap plane =
		match plane with
			| LIGHT -> lightMap;
			| DARK -> darkMap;
end;;
