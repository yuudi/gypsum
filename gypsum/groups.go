package gypsum

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Item struct {
	ItemType ItemType `json:"item_type"`
	ItemID   uint64   `json:"item_id"`
}

type Group struct {
	DisplayName   string `json:"display_name"`
	PluginName    string `json:"plugin_name"`
	PluginVersion int64  `json:"plugin_version"`
	Items         []Item `json:"items"`
	ParentGroup   uint64 `json:"-"`
}

type ArchiveItem struct {
	ItemType  ItemType
	ItemBytes []byte
}

type GroupArchive struct {
	DisplayName   string
	PluginName    string
	PluginVersion int64
	GypsumVersion string
	ArchiveItems  []ArchiveItem
}

var groups map[uint64]*Group

func (g *Group) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(g); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func GroupFromBytes(b []byte) (*Group, error) {
	g := &Group{
		DisplayName: "",
		Items:       []Item{},
		ParentGroup: 0,
	}
	buffer := bytes.Buffer{}
	buffer.Write(b)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(g)
	return g, err
}

func (g *GroupArchive) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(g); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (g Group) ExportToArchive(name string, version int64) *GroupArchive {
	archiveItems := make([]ArchiveItem, len(g.Items))
	for i, item := range g.Items {
		it, ok := findItem(item.ItemType, item.ItemID)
		if !ok {
			log.Errorf("cannot find item: type:%s, id: %d", item.ItemType, item.ItemID)
			continue
		}
		itBytes, err := it.ToBytes()
		if err != nil {
			log.Error(err)
			continue
		}
		archiveItems[i] = ArchiveItem{
			ItemType:  item.ItemType,
			ItemBytes: itBytes,
		}
	}
	return &GroupArchive{
		DisplayName:   g.DisplayName,
		PluginName:    name,
		PluginVersion: version,
		GypsumVersion: os.Getenv("GYPSUM_VERSION"),
		ArchiveItems:  archiveItems,
	}
}

func GroupFromArchiveReader(reader io.Reader) (*Group, error) {
	ga := &GroupArchive{
		DisplayName:   "",
		PluginName:    "",
		PluginVersion: 0,
		GypsumVersion: "",
		ArchiveItems:  nil,
	}
	decoder := gob.NewDecoder(reader)
	if err := decoder.Decode(ga); err != nil {
		log.Debug("02011755")
		return nil, err
	}
	g := &Group{
		DisplayName:   ga.DisplayName,
		PluginName:    ga.PluginName,
		PluginVersion: ga.PluginVersion,
		Items:         nil,
		ParentGroup:   0,
	}
	g.Items = make([]Item, len(ga.ArchiveItems))
	for i, item := range ga.ArchiveItems {
		idx, err := RestoreFromUserRecord(item.ItemType, item.ItemBytes)
		if err != nil {
			log.Error(err)
			continue
		}
		g.Items[i] = Item{
			ItemType: item.ItemType,
			ItemID:   idx,
		}
	}
	return g, nil
}

func loadGroups() {
	groups = make(map[uint64]*Group)
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-groups-")), nil)
	defer func() {
		iter.Release()
		if err := iter.Error(); err != nil {
			log.Errorf("载入数据错误：%s", err)
		}
	}()
	for iter.Next() {
		key := ToUint(iter.Key()[14:])
		value := iter.Value()
		g, e := GroupFromBytes(value)
		if e != nil {
			log.Errorf("无法加载组%d：%s", key, e)
			continue
		}
		groups[key] = g
	}
}

func (g *Group) SaveToDB(gid uint64) error {
	if gid == 0 {
		return nil
	}
	v, err := g.ToBytes()
	if err != nil {
		return err
	}
	return db.Put(append([]byte("gypsum-groups-"), U64ToBytes(gid)...), v, nil)
}

func findItem(itemType ItemType, itemID uint64) (item UserRecord, ok bool) {
	switch itemType {
	case RuleItem:
		item, ok = rules[itemID]
	case TriggerItem:
		item, ok = triggers[itemID]
	case SchedulerItem:
		item, ok = jobs[itemID]
	case ResourceItem:
		item, ok = resources[itemID]
	case GroupItem:
		item, ok = groups[itemID]
	default:
		ok = false
	}
	return
}

func (g *Group) GetParentID() uint64 {
	return g.ParentGroup
}

func (g *Group) NewParent(selfID, parentID uint64) error {
	g.ParentGroup = parentID
	err := g.SaveToDB(selfID)
	return err
}

func getGroups(c *gin.Context) {
	c.JSON(200, groups)
}

func getGroupByID(c *gin.Context) {
	groupIDStr := c.Param("gid")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such group",
		})
		return
	}
	g, ok := groups[groupID]
	if ok {
		c.JSON(200, g)
		return
	}
	c.JSON(404, gin.H{
		"code":    1000,
		"message": "no such group",
	})
}

func createGroup(c *gin.Context) {
	if c.ContentType() == "application/zip" {
		importGroup(c)
		return
	}
	var group Group
	if err := c.BindJSON(&group); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	parentStr := c.Param("gid")
	var parentID uint64
	if len(parentStr) == 0 {
		parentID = 0
	} else {
		var err error
		parentID, err = strconv.ParseUint(parentStr, 10, 64)
		if err != nil {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such group",
			})
			return
		}
		if parentID != 0 {
			c.JSON(400, gin.H{
				"code":    1400,
				"message": "group in group are not supported yet",
			})
			return
		}
	}
	group.ParentGroup = parentID

	cursor++
	if err := db.Put([]byte("gypsum-$meta-cursor"), U64ToBytes(cursor), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3031,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	err := group.SaveToDB(cursor)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    3032,
			"message": fmt.Sprintf("Server got itself into trouble: item does not exist in its parent group"),
		})
		return
	}
	groups[cursor] = &group
	c.JSON(201, gin.H{
		"code":     0,
		"message":  "ok",
		"group_id": cursor,
	})
	return
}

func addGroupItem(c *gin.Context) {
	itemIDStr := c.Param("iid")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such item",
		})
		return
	}
	groupIDStr := c.Param("gid")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such item",
		})
		return
	}
	group, ok := groups[groupID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1001,
			"message": "no such group",
		})
		return
	}
	var item UserRecord
	iType := ItemType(c.Param("type"))
	if iType == GroupItem {
		c.JSON(422, gin.H{
			"code":    3010,
			"message": "not supported yet",
		})
		return
	}
	item, ok = findItem(iType, itemID)
	if !ok {
		c.JSON(404, gin.H{
			"code":    1002,
			"message": "item not found",
		})
		return
	}
	oldGroupID := item.GetParentID()
	if oldGroupID != 0 {
		// remove item from old group
		oldGroup := groups[oldGroupID]
		for k, v := range oldGroup.Items {
			if v.ItemID == itemID {
				// remove the k-th element in a slice
				copy(oldGroup.Items[k:], oldGroup.Items[k+1:])
				oldGroup.Items = oldGroup.Items[:len(oldGroup.Items)-1]
				break
			}
			log.Error("item does not exist in its parent group")
			c.JSON(500, gin.H{
				"code":    3051,
				"message": fmt.Sprintf("Server got itself into trouble: item does not exist in its parent group"),
			})
			return
		}
		if err := oldGroup.SaveToDB(oldGroupID); err != nil {
			c.JSON(500, gin.H{
				"code":    3052,
				"message": fmt.Sprintf("Server got itself into trouble: item does not exist in its parent group"),
			})
			return
		}
	}
	if err = item.NewParent(itemID, groupID); err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	group.Items = append(group.Items, Item{
		ItemType: iType,
		ItemID:   itemID,
	})
	if err := group.SaveToDB(groupID); err != nil {
		c.JSON(500, gin.H{
			"code":    3053,
			"message": fmt.Sprintf("Server got itself into trouble: item does not exist in its parent group"),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
}

func exportGroup(c *gin.Context) {
	pluginName := c.Query("plugin_name")
	pluginVersionStr := c.Query("plugin_version")
	groupIDStr := c.Param("gid")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such group",
		})
		return
	}
	pluginVersion, err := strconv.ParseInt(pluginVersionStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    1080,
			"message": "plugin_version must be an integer",
		})
		return
	}
	group, ok := groups[groupID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such group",
		})
		return
	}
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	f, err := zipWriter.Create("gypsum-plugin.dat")
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    7021,
			"message": fmt.Sprintf("error when create plugin zipfile: %s", err),
		})
		return
	}
	groupData, err := group.ExportToArchive(pluginName, pluginVersion).ToBytes()
	if err != nil {
		c.JSON(500, gin.H{
			"code":    3041,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	_, err = f.Write(groupData)
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    3042,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	for _, item := range group.Items {
		if item.ItemType == ResourceItem {
			// attach all resources
			res := resources[item.ItemID]
			fileData, err := os.ReadFile(path.Join(resDir, res.Sha256Sum+res.Ext))
			if err != nil {
				log.Error(err)
				c.JSON(500, gin.H{
					"code":    3043,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			f, err := zipWriter.Create(res.Sha256Sum + res.Ext)
			if err != nil {
				log.Error(err)
				c.JSON(500, gin.H{
					"code":    3044,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			_, err = f.Write(fileData)
			if err != nil {
				log.Error(err)
				c.JSON(500, gin.H{
					"code":    3045,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
		}
	}
	err = zipWriter.Close()
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    3046,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+ReplaceFilename(pluginName, "_")+".zip")
	c.Header("Content-Type", "application/octet-stream")
	_, err = c.Writer.Write(buf.Bytes())
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    3047,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
}

func importGroup(c *gin.Context) {
	if c.ContentType() != "application/zip" {
		c.JSON(415, gin.H{
			"code":    5000,
			"message": fmt.Sprintf("request type do not meet application/zip: %s", c.ContentType()),
		})
	}
	bodyReader := c.Request.Body
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    6000,
			"message": fmt.Sprintf("error when reading request body: %s", err),
		})
		return
	}
	zipReader, err := zip.NewReader(bytes.NewReader(body), c.Request.ContentLength)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    5000,
			"message": fmt.Sprintf("cannot read body as zipfile: %s", err),
		})
	}
	var newGroup *Group
	for _, file := range zipReader.File {
		if file.Name == "gypsum-plugin.dat" {
			fr, err := file.Open()
			if err != nil {
				log.Error(err)
				c.JSON(500, gin.H{
					"code":    3000,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			newGroup, err = GroupFromArchiveReader(fr)
			if err != nil {
				log.Error(err)
				c.JSON(500, gin.H{
					"code":    3000,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			continue
		}
		nameSplit := strings.Split(file.Name, ".")
		if len(nameSplit[0]) == 64 {
			_, exists := resourceIDByHash(nameSplit[0])
			if exists {
				continue
			}
			fr, err := file.Open()
			if err != nil {
				log.Error(err)
				c.JSON(500, gin.H{
					"code":    3000,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			body, err := io.ReadAll(fr)
			if err != nil {
				log.Error(err)
				c.JSON(500, gin.H{
					"code":    3000,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			hashBytes := sha256.Sum256(body)
			hashHex := hex.EncodeToString(hashBytes[:])
			if !strings.EqualFold(nameSplit[0], hashHex) {
				c.JSON(400, gin.H{
					"code":    3000,
					"message": fmt.Sprintf("zipfile sha256sum dose not match fine name: %s", file.Name),
				})
				return
			}
			if err := os.WriteFile(path.Join(resDir, file.Name), body, 0444); err != nil {
				c.JSON(500, gin.H{
					"code":    6000,
					"message": fmt.Sprintf("error when writing file: %s", err),
				})
				return
			}
		}
	}
	if newGroup == nil {
		c.JSON(412, gin.H{
			"code":    4000,
			"message": fmt.Sprintf("zipfile has no gypsum metadata"),
		})
		return
	}
	cursor++
	if err := db.Put([]byte("gypsum-$meta-cursor"), U64ToBytes(cursor), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	groups[cursor] = newGroup
	if err = newGroup.SaveToDB(cursor); err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	c.JSON(201, gin.H{
		"code":     0,
		"message":  "ok",
		"group_id": cursor,
	})
}

type groupMoveTo struct {
	MoveTo uint64 `json:"move_to"`
}

func deleteGroup(c *gin.Context) {
	groupIDStr := c.Param("gid")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such group",
		})
		return
	}
	group, ok := groups[groupID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	movePatch := groupMoveTo{}
	if err := c.BindJSON(movePatch); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	var newGroup *Group
	if movePatch.MoveTo != 0 {
		newGroup, ok = groups[movePatch.MoveTo]
		if !ok {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such group",
			})
			return
		}
	} else {
		newGroup = nil
	}
	for _, item := range group.Items {
		it, ok := findItem(item.ItemType, item.ItemID)
		if !ok {
			log.Errorf("cannot find item: type:%s, id: %d", item.ItemType, item.ItemID)
			continue
		}
		err = it.NewParent(item.ItemID, movePatch.MoveTo)
		if err != nil {
			log.Error(err)
			continue
		}
	}
	if newGroup != nil {
		newGroup.Items = append(newGroup.Items, group.Items...)
	}
	delete(groups, groupID)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "deleted",
	})
	return
}

type groupNamePatch struct {
	DisplayName string `json:"display_name"`
}

func renameGroup(c *gin.Context) {
	groupIDStr := c.Param("rid")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	group, ok := groups[groupID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such group",
		})
		return
	}
	np := groupNamePatch{}
	if err := c.BindJSON(&np); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	group.DisplayName = np.DisplayName
	err = group.SaveToDB(groupID)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
}
