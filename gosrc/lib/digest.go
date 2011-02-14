package lightwave

func countBlips(blips []interface{}) int {
  count := len(blips)
  for _, x := range blips {
    for _, t := range getArray(getObject(x)["threads"]) {
      count += countBlips( getArray(getObject(t)["blips"]) )
    }
  }
  return count
}

func getObject(obj interface{}) map[string]interface{} {
  result, ok := obj.(map[string]interface{})
  if ok {
    return result
  }
  return make(map[string]interface{})
}

func getArray(obj interface{}) []interface{} {
  result, ok := obj.([]interface{})
  if ok {
    return result
  }
  return []interface{}{}
}

func getString(obj interface{}) string {
  result, ok := obj.(string)
  if ok {
    return result
  }
  return ""
}

func GetTags(node *DocumentNode) []string {
  switch node.Schema() {
  case "//lightwave/blips":
    return BlipsTags(node)
  case "//lightwave/user":
    return UserTags(node)
  case "//lightwave/friend-request":
    return FriendRequestTags(node)
  }
  return []string{}
}

func GetDigest(node *DocumentNode, mapping DocumentMappingId) map[string]interface{} {
  switch node.Schema() {
  case "//lightwave/blips":
    return BlipsDigest(node)
  case "//lightwave/user":
    return UserDigest(node)
  case "//lightwave/friend-request":
    return FriendRequestDigest(node)
  }
  return make(map[string]interface{})
}

func BlipsTags(node *DocumentNode) []string {
  users := node.Participants()  
  // Compute tags for the indexer
  tags := make([]string, 0, len(users))
  for _, u := range users {
    tags = append(tags, "with:" + u.Username + "@" + u.Domain)
  }
  return tags
}

func UserTags(node *DocumentNode) []string {
  friends := getArray(getObject(getObject(node.doc)["_data"])["friends"])
  // Compute tags for the indexer
  tags := make([]string, 0, len(friends))
  for _, f := range friends {
    tags = append(tags, "friend:" + getString(f))
  }
  tags = append(tags, "schema://lightwave/user")
  return tags
}

func FriendRequestTags(node *DocumentNode) []string {
  users := node.Participants()  
  // Compute tags for the indexer
  tags := make([]string, 0, len(users))
  for _, u := range users {
    tags = append(tags, "with:" + u.Username + "@" + u.Domain)
  }
  request := getObject(getObject(getObject(node.doc)["_data"])["request"])
  response := getObject(getObject(getObject(node.doc)["_data"])["response"])
  tags = append(tags, "friend-req:" + getString(request["userid"]))
  tags = append(tags, "friend-res:" + getString(response["userid"]))
  tags = append(tags, "friend-state:" + getString(response["state"]))

  return tags
}

func BlipsDigest(node *DocumentNode) map[string]interface{} {
  result := make(map[string]interface{})
  blips := getArray(getObject(node.doc["_data"])["blips"])
  if len(blips) > 0 {
    result["author"] = getString(getObject(getObject(blips[0])["_meta"])["author"])
    text := getArray(getObject(getObject(blips[0])["content"])["text"])
    digest := ""
    for _, t := range text {
      if str, ok := t.(string); ok {
	digest += str
      } else if getString(getObject(t)["_type"]) == "parag" {
	if digest != "" {
	  digest += "</br>"
	}
      }	
    }
    result["text"] = digest
  }
  // Add data about the last 5 comments in the main thread
  comments := make([]interface{}, 0, 3)
  l := len(blips) - 1
  if l > 3 {
    l = 3
  }
  for i := len(blips) - l; i < len(blips); i++ {
    c := make(map[string]interface{})
    c["author"] = getString(getObject(getObject(blips[i])["_meta"])["author"])
    text := getArray(getObject(getObject(blips[i])["content"])["text"])
    digest := ""
    for _, t := range text {
      if str, ok := t.(string); ok {
	digest += str
      } else if getString(getObject(t)["_type"]) == "parag" {
	if digest != "" {
	  digest += "</br>"
	}
      }	
    }
    c["text"] = digest
    comments = append(comments, c)
  }
  result["comments"] = comments
  result["blipCount"] = countBlips(blips)
  return result
}

func UserDigest(node *DocumentNode) map[string]interface{} {
  result := make(map[string]interface{})
  data := getObject(node.doc["_data"])
  result["userid"] = data["userid"]
  result["displayName"] = data["displayName"]
  result["image"] = data["image"]
  return result 
}

func FriendRequestDigest(node *DocumentNode) map[string]interface{} {
  result := BlipsDigest(node)
  request := getObject(getObject(getObject(node.doc)["_data"])["request"])
  response := getObject(getObject(getObject(node.doc)["_data"])["response"])
  r := make(map[string]interface{})
  r["userid"] = getString(request["userid"])
  result["request"] = r
  r = make(map[string]interface{})
  r["userid"] = getString(response["userid"])
  r["state"] = getString(response["state"])
  result["response"] = r
  result["schema"] = "//lightwave/friend-request"
  return result
}