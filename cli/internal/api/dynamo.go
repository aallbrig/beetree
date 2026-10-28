package api

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Single-table layout (all keys are strings):
//
//	TREE#<id>    META             tree summary (data), spec, likeCount, commentCount
//	TREE#<id>    LIKE#<user>      one like
//	TREE#<id>    COMMENT#<ts>#<id>
//	USER#<id>    PROFILE
//	USER#<id>    STATS            forum counters
//	THREAD#<id>  META             thread (data), postCount, lastActivity
//	THREAD#<id>  POST#<ts>#<id>
//	SCENE#<id>   META
//
// gsi1: owner index (gsi1pk=OWNER#<id>, gsi1sk=TREE#<ts> | SCENE#<ts>)
// gsi2: feeds       (gsi2pk=PUBLIC#TREE | FORUM, gsi2sk=<ts>#<id>)
//
// Mutable counters live in their own attributes so they can be updated
// atomically without rewriting the JSON document.
const (
	gsiOwner  = "gsi1"
	gsiFeed   = "gsi2"
	feedTree  = "PUBLIC#TREE"
	feedForum = "FORUM"
)

// DynamoAPI is the subset of the DynamoDB client the store uses.
type DynamoAPI interface {
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, in *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	BatchWriteItem(ctx context.Context, in *dynamodb.BatchWriteItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	TransactWriteItems(ctx context.Context, in *dynamodb.TransactWriteItemsInput, opts ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// DynamoStore is the production Store.
type DynamoStore struct {
	DB    DynamoAPI
	Table string
}

type item = map[string]types.AttributeValue

func s(v string) types.AttributeValue { return &types.AttributeValueMemberS{Value: v} }
func n(v int) types.AttributeValue    { return &types.AttributeValueMemberN{Value: strconv.Itoa(v)} }

func key(pk, sk string) item { return item{"pk": s(pk), "sk": s(sk)} }

func str(it item, name string) string {
	if v, ok := it[name].(*types.AttributeValueMemberS); ok {
		return v.Value
	}
	return ""
}

func num(it item, name string) int {
	if v, ok := it[name].(*types.AttributeValueMemberN); ok {
		i, _ := strconv.Atoi(v.Value)
		return i
	}
	return 0
}

func (d *DynamoStore) get(ctx context.Context, pk, sk string) (item, error) {
	out, err := d.DB.GetItem(ctx, &dynamodb.GetItemInput{TableName: &d.Table, Key: key(pk, sk), ConsistentRead: aws.Bool(true)})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, ErrNotFound
	}
	return out.Item, nil
}

func (d *DynamoStore) queryAll(ctx context.Context, in *dynamodb.QueryInput) ([]item, error) {
	in.TableName = &d.Table
	var all []item
	for {
		out, err := d.DB.Query(ctx, in)
		if err != nil {
			return nil, err
		}
		all = append(all, out.Items...)
		if out.LastEvaluatedKey == nil || (in.Limit != nil && len(all) >= int(*in.Limit)) {
			return all, nil
		}
		in.ExclusiveStartKey = out.LastEvaluatedKey
	}
}

func (d *DynamoStore) listPartition(ctx context.Context, pk, skPrefix string) ([]item, error) {
	return d.queryAll(ctx, &dynamodb.QueryInput{
		KeyConditionExpression:    aws.String("pk = :pk AND begins_with(sk, :p)"),
		ExpressionAttributeValues: item{":pk": s(pk), ":p": s(skPrefix)},
		ConsistentRead:            aws.Bool(true),
	})
}

func (d *DynamoStore) queryIndex(ctx context.Context, index, pkAttr, skAttr, pk, skPrefix string, newestFirst bool, limit int) ([]item, error) {
	cond := "#pk = :pk"
	vals := item{":pk": s(pk)}
	if skPrefix != "" {
		cond += " AND begins_with(#sk, :p)"
		vals[":p"] = s(skPrefix)
	}
	in := &dynamodb.QueryInput{
		IndexName:                 aws.String(index),
		KeyConditionExpression:    aws.String(cond),
		ExpressionAttributeNames:  map[string]string{"#pk": pkAttr, "#sk": skAttr},
		ExpressionAttributeValues: vals,
		ScanIndexForward:          aws.Bool(!newestFirst),
	}
	if limit > 0 {
		in.Limit = aws.Int32(int32(min(limit, 1000))) //nolint:gosec // bounded to 1000 on the line above
	}
	return d.queryAll(ctx, in)
}

// deletePartition removes every item under pk (used when deleting a tree).
func (d *DynamoStore) deletePartition(ctx context.Context, pk string) error {
	items, err := d.queryAll(ctx, &dynamodb.QueryInput{
		KeyConditionExpression:    aws.String("pk = :pk"),
		ExpressionAttributeValues: item{":pk": s(pk)},
		ProjectionExpression:      aws.String("pk, sk"),
		ConsistentRead:            aws.Bool(true),
	})
	if err != nil {
		return err
	}
	for len(items) > 0 {
		batch := items
		if len(batch) > 25 {
			batch = batch[:25]
		}
		items = items[len(batch):]
		reqs := make([]types.WriteRequest, len(batch))
		for i, it := range batch {
			reqs[i] = types.WriteRequest{DeleteRequest: &types.DeleteRequest{Key: key(str(it, "pk"), str(it, "sk"))}}
		}
		if _, err := d.DB.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{RequestItems: map[string][]types.WriteRequest{d.Table: reqs}}); err != nil {
			return err
		}
	}
	return nil
}

func isCancelled(err error) bool {
	var tc *types.TransactionCanceledException
	var cc *types.ConditionalCheckFailedException
	return errors.As(err, &tc) || errors.As(err, &cc)
}

// ---- trees -----------------------------------------------------------------

func treeFromItem(it item) (Tree, error) {
	var t Tree
	if err := json.Unmarshal([]byte(str(it, "data")), &t); err != nil {
		return t, err
	}
	t.Spec = str(it, "spec")
	t.LikeCount, t.CommentCount = num(it, "likeCount"), num(it, "commentCount")
	return t, nil
}

func (d *DynamoStore) PutTree(ctx context.Context, t Tree) error {
	spec := t.Spec
	t.Spec, t.LikeCount, t.CommentCount, t.LikedByMe = "", 0, 0, false
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	set := "SET #data = :d, spec = :spec, gsi1pk = :g1p, gsi1sk = :g1s, likeCount = if_not_exists(likeCount, :z), commentCount = if_not_exists(commentCount, :z)"
	vals := item{":d": s(string(data)), ":spec": s(spec), ":g1p": s("OWNER#" + t.OwnerID), ":g1s": s("TREE#" + t.CreatedAt + "#" + t.ID), ":z": n(0)}
	if t.Visibility == VisibilityPublic {
		set += ", gsi2pk = :g2p, gsi2sk = :g2s"
		vals[":g2p"], vals[":g2s"] = s(feedTree), s(t.CreatedAt+"#"+t.ID)
	} else {
		set += " REMOVE gsi2pk, gsi2sk"
	}
	_, err = d.DB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &d.Table, Key: key("TREE#"+t.ID, "META"), UpdateExpression: &set,
		ExpressionAttributeNames: map[string]string{"#data": "data"}, ExpressionAttributeValues: vals,
	})
	return err
}

func (d *DynamoStore) GetTree(ctx context.Context, id string) (Tree, error) {
	it, err := d.get(ctx, "TREE#"+id, "META")
	if err != nil {
		return Tree{}, err
	}
	return treeFromItem(it)
}

func (d *DynamoStore) DeleteTree(ctx context.Context, id string) error {
	if _, err := d.get(ctx, "TREE#"+id, "META"); err != nil {
		return err
	}
	return d.deletePartition(ctx, "TREE#"+id)
}

func (d *DynamoStore) treesFrom(items []item) ([]Tree, error) {
	out := make([]Tree, 0, len(items))
	for _, it := range items {
		t, err := treeFromItem(it)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (d *DynamoStore) ListPublicTrees(ctx context.Context, limit int) ([]Tree, error) {
	items, err := d.queryIndex(ctx, gsiFeed, "gsi2pk", "gsi2sk", feedTree, "", true, limit)
	if err != nil {
		return nil, err
	}
	return d.treesFrom(items)
}

func (d *DynamoStore) ListTreesByOwner(ctx context.Context, ownerID string) ([]Tree, error) {
	items, err := d.queryIndex(ctx, gsiOwner, "gsi1pk", "gsi1sk", "OWNER#"+ownerID, "TREE#", true, 0)
	if err != nil {
		return nil, err
	}
	return d.treesFrom(items)
}

// ---- likes & comments ------------------------------------------------------

func (d *DynamoStore) HasLiked(ctx context.Context, treeID, userID string) (bool, error) {
	_, err := d.get(ctx, "TREE#"+treeID, "LIKE#"+userID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (d *DynamoStore) SetLike(ctx context.Context, treeID, userID string, liked bool) (int, error) {
	tree, err := d.GetTree(ctx, treeID)
	if err != nil {
		return 0, err
	}
	has, err := d.HasLiked(ctx, treeID, userID)
	if err != nil {
		return 0, err
	}
	if has == liked {
		return tree.LikeCount, nil
	}
	treeKey := key("TREE#"+treeID, "META")
	likeKey := key("TREE#"+treeID, "LIKE#"+userID)
	delta := "1"
	var op types.TransactWriteItem
	if liked {
		op = types.TransactWriteItem{Put: &types.Put{TableName: &d.Table, Item: likeKey, ConditionExpression: aws.String("attribute_not_exists(pk)")}}
	} else {
		delta = "-1"
		op = types.TransactWriteItem{Delete: &types.Delete{TableName: &d.Table, Key: likeKey, ConditionExpression: aws.String("attribute_exists(pk)")}}
	}
	_, err = d.DB.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
		op,
		{Update: &types.Update{TableName: &d.Table, Key: treeKey, UpdateExpression: aws.String("ADD likeCount :d"),
			ExpressionAttributeValues: item{":d": &types.AttributeValueMemberN{Value: delta}}}},
	}})
	if err != nil && !isCancelled(err) { // cancelled = lost a race with another identical request
		return 0, err
	}
	tree, err = d.GetTree(ctx, treeID)
	return tree.LikeCount, err
}

func (d *DynamoStore) AddComment(ctx context.Context, c Comment) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	it := key("TREE#"+c.TreeID, "COMMENT#"+c.CreatedAt+"#"+c.ID)
	it["data"], it["cid"] = s(string(data)), s(c.ID)
	_, err = d.DB.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
		{Put: &types.Put{TableName: &d.Table, Item: it}},
		{Update: &types.Update{TableName: &d.Table, Key: key("TREE#"+c.TreeID, "META"), UpdateExpression: aws.String("ADD commentCount :one"),
			ConditionExpression: aws.String("attribute_exists(pk)"), ExpressionAttributeValues: item{":one": n(1)}}},
	}})
	if isCancelled(err) {
		return ErrNotFound
	}
	return err
}

func (d *DynamoStore) commentItems(ctx context.Context, treeID string) ([]item, error) {
	return d.listPartition(ctx, "TREE#"+treeID, "COMMENT#")
}

func (d *DynamoStore) ListComments(ctx context.Context, treeID string) ([]Comment, error) {
	items, err := d.commentItems(ctx, treeID)
	if err != nil {
		return nil, err
	}
	out := make([]Comment, 0, len(items))
	for _, it := range items {
		var c Comment
		if err := json.Unmarshal([]byte(str(it, "data")), &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (d *DynamoStore) findComment(ctx context.Context, treeID, commentID string) (item, error) {
	items, err := d.commentItems(ctx, treeID)
	if err != nil {
		return nil, err
	}
	for _, it := range items {
		if str(it, "cid") == commentID {
			return it, nil
		}
	}
	return nil, ErrNotFound
}

func (d *DynamoStore) GetComment(ctx context.Context, treeID, commentID string) (Comment, error) {
	it, err := d.findComment(ctx, treeID, commentID)
	if err != nil {
		return Comment{}, err
	}
	var c Comment
	return c, json.Unmarshal([]byte(str(it, "data")), &c)
}

func (d *DynamoStore) DeleteComment(ctx context.Context, treeID, commentID string) error {
	it, err := d.findComment(ctx, treeID, commentID)
	if err != nil {
		return err
	}
	_, err = d.DB.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
		{Delete: &types.Delete{TableName: &d.Table, Key: key(str(it, "pk"), str(it, "sk")), ConditionExpression: aws.String("attribute_exists(pk)")}},
		{Update: &types.Update{TableName: &d.Table, Key: key("TREE#"+treeID, "META"), UpdateExpression: aws.String("ADD commentCount :neg"),
			ExpressionAttributeValues: item{":neg": n(-1)}}},
	}})
	if isCancelled(err) {
		return ErrNotFound
	}
	return err
}

// ---- profiles --------------------------------------------------------------

func (d *DynamoStore) PutProfile(ctx context.Context, p Profile) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	it := key("USER#"+p.UserID, "PROFILE")
	it["data"] = s(string(data))
	_, err = d.DB.PutItem(ctx, &dynamodb.PutItemInput{TableName: &d.Table, Item: it})
	return err
}

func (d *DynamoStore) GetProfile(ctx context.Context, userID string) (Profile, error) {
	it, err := d.get(ctx, "USER#"+userID, "PROFILE")
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	return p, json.Unmarshal([]byte(str(it, "data")), &p)
}

// ---- forum -----------------------------------------------------------------

func threadFromItem(it item) (Thread, error) {
	var t Thread
	if err := json.Unmarshal([]byte(str(it, "data")), &t); err != nil {
		return t, err
	}
	t.PostCount = num(it, "postCount")
	if la := str(it, "lastActivity"); la != "" {
		t.LastActivity = la
	}
	return t, nil
}

func (d *DynamoStore) bumpUserStat(ctx context.Context, userID, attr string) error {
	_, err := d.DB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &d.Table, Key: key("USER#"+userID, "STATS"),
		UpdateExpression:          aws.String("ADD " + attr + " :one"),
		ExpressionAttributeValues: item{":one": n(1)},
	})
	return err
}

func (d *DynamoStore) PutThread(ctx context.Context, t Thread) error {
	last := t.LastActivity
	t.PostCount, t.LastActivity = 0, ""
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	out, err := d.DB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &d.Table, Key: key("THREAD#"+t.ID, "META"),
		UpdateExpression:          aws.String("SET #data = :d, lastActivity = if_not_exists(lastActivity, :la), gsi2pk = :g2p, gsi2sk = if_not_exists(gsi2sk, :g2s), postCount = if_not_exists(postCount, :z)"),
		ExpressionAttributeNames:  map[string]string{"#data": "data"},
		ExpressionAttributeValues: item{":d": s(string(data)), ":la": s(last), ":g2p": s(feedForum), ":g2s": s(last + "#" + t.ID), ":z": n(0)},
		ReturnValues:              types.ReturnValueUpdatedOld,
	})
	if err != nil {
		return err
	}
	if len(out.Attributes) == 0 { // brand-new thread
		return d.bumpUserStat(ctx, t.AuthorID, "threads")
	}
	return nil
}

func (d *DynamoStore) GetThread(ctx context.Context, id string) (Thread, error) {
	it, err := d.get(ctx, "THREAD#"+id, "META")
	if err != nil {
		return Thread{}, err
	}
	return threadFromItem(it)
}

func (d *DynamoStore) ListThreads(ctx context.Context, limit int) ([]Thread, error) {
	items, err := d.queryIndex(ctx, gsiFeed, "gsi2pk", "gsi2sk", feedForum, "", true, limit)
	if err != nil {
		return nil, err
	}
	out := make([]Thread, 0, len(items))
	for _, it := range items {
		t, err := threadFromItem(it)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (d *DynamoStore) AddPost(ctx context.Context, p Post) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	it := key("THREAD#"+p.ThreadID, "POST#"+p.CreatedAt+"#"+p.ID)
	it["data"] = s(string(data))
	_, err = d.DB.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
		{Put: &types.Put{TableName: &d.Table, Item: it}},
		{Update: &types.Update{TableName: &d.Table, Key: key("THREAD#"+p.ThreadID, "META"),
			UpdateExpression:          aws.String("ADD postCount :one SET lastActivity = :la, gsi2sk = :g2s"),
			ConditionExpression:       aws.String("attribute_exists(pk)"),
			ExpressionAttributeValues: item{":one": n(1), ":la": s(p.CreatedAt), ":g2s": s(p.CreatedAt + "#" + p.ThreadID)}}},
	}})
	if isCancelled(err) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return d.bumpUserStat(ctx, p.AuthorID, "posts")
}

func (d *DynamoStore) ListPosts(ctx context.Context, threadID string) ([]Post, error) {
	items, err := d.listPartition(ctx, "THREAD#"+threadID, "POST#")
	if err != nil {
		return nil, err
	}
	out := make([]Post, 0, len(items))
	for _, it := range items {
		var p Post
		if err := json.Unmarshal([]byte(str(it, "data")), &p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// ---- scenes ----------------------------------------------------------------

func (d *DynamoStore) PutScene(ctx context.Context, sc Scene) error {
	data, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	it := key("SCENE#"+sc.ID, "META")
	it["data"] = s(string(data))
	it["gsi1pk"], it["gsi1sk"] = s("OWNER#"+sc.OwnerID), s("SCENE#"+sc.CreatedAt+"#"+sc.ID)
	_, err = d.DB.PutItem(ctx, &dynamodb.PutItemInput{TableName: &d.Table, Item: it})
	return err
}

func (d *DynamoStore) GetScene(ctx context.Context, id string) (Scene, error) {
	it, err := d.get(ctx, "SCENE#"+id, "META")
	if err != nil {
		return Scene{}, err
	}
	var sc Scene
	return sc, json.Unmarshal([]byte(str(it, "data")), &sc)
}

func (d *DynamoStore) DeleteScene(ctx context.Context, id string) error {
	_, err := d.DB.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &d.Table, Key: key("SCENE#"+id, "META"), ConditionExpression: aws.String("attribute_exists(pk)")})
	if isCancelled(err) {
		return ErrNotFound
	}
	return err
}

func (d *DynamoStore) ListScenesByOwner(ctx context.Context, ownerID string) ([]Scene, error) {
	items, err := d.queryIndex(ctx, gsiOwner, "gsi1pk", "gsi1sk", "OWNER#"+ownerID, "SCENE#", true, 0)
	if err != nil {
		return nil, err
	}
	out := make([]Scene, 0, len(items))
	for _, it := range items {
		var sc Scene
		if err := json.Unmarshal([]byte(str(it, "data")), &sc); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, nil
}

// ---- stats -----------------------------------------------------------------

func (d *DynamoStore) Stats(ctx context.Context, userID string) (Stats, error) {
	var st Stats
	trees, err := d.ListTreesByOwner(ctx, userID)
	if err != nil {
		return st, err
	}
	for _, t := range trees {
		st.Trees++
		if t.Visibility == VisibilityPublic {
			st.PublicTrees++
		}
		st.Likes += t.LikeCount
		st.Comments += t.CommentCount
	}
	scenes, err := d.ListScenesByOwner(ctx, userID)
	if err != nil {
		return st, err
	}
	st.ScenesCreated = len(scenes)
	if it, err := d.get(ctx, "USER#"+userID, "STATS"); err == nil {
		st.ForumThreads, st.ForumPosts = num(it, "threads"), num(it, "posts")
	} else if !errors.Is(err, ErrNotFound) {
		return st, err
	}
	return st, nil
}
