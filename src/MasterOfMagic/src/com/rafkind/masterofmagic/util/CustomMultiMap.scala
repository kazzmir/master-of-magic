/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package com.rafkind.masterofmagic.util
import scala.collection.mutable._;

class CustomMultiMap[KeyType, ValueType] {
  var data = new HashMap[KeyType, Set[ValueType]];

  def get(k:KeyType):Option[Set[ValueType]] = 
    data.get(k);

  def put(k:KeyType, v:ValueType):CustomMultiMap[KeyType, ValueType] = {
    data.get(k) match {
      case Some(set) => 
        set.add(v);
      case None => 
        var set = new HashSet[ValueType];
        set.add(v);
        data += k -> set;
    }

    return this;
  }
}
