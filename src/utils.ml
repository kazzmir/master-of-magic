let inject func times =
  ExtList.List.of_enum (Enum.map func (ExtList.List.enum
  (ExtList.List.make times 0)))
;;
