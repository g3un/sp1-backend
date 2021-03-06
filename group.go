package main

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type member struct {
	AppId   string `json:"app_id" bson:"app_id"`
	IsAdmin int    `json:"is_admin" bson:"is_admin"`
}

type group struct {
	GroupId    primitive.ObjectID `json:"group_id" bson:"_id,omitempty"`
	Ott        string             `json:"ott" bson:"ott"`
	Account    account            `json:"account" bson:"account"`
	UpdateTime int64              `json:"update_time" bson:"update_time"`
	Members    []member           `json:"members" bson:"members"`
}

func getGroup(c *fiber.Ctx) error {
	client, ctx, cancel, err := newClient()
	if err != nil {
		return err
	}
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(c.Params("groupId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	filter := bson.M{"_id": _id}

	num, err := getCollection(client, "group").CountDocuments(ctx, filter)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var group group
	if num != 1 {
		return fiber.ErrNotFound
	}

	if err = getCollection(client, "group").FindOne(ctx, filter).Decode(&group); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	groupByte, err := sonic.Marshal(group)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Send(groupByte)
}

func postGroup(c *fiber.Ctx) error {
	client, ctx, cancel, err := newClient()
	if err != nil {
		return err
	}
	defer cancel()

	var parser struct {
		AppId string `json:"app_id" bson:"app_id"`
		Ott   string `json:"ott" bson:"ott"`
		OttId string `json:"ott_id" bson:"ott_id"`
		OttPw string `json:"ott_pw" bson:"ott_pw"`
	}
	if err = c.BodyParser(&parser); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if parser.AppId == "" || parser.Ott == "" || parser.OttId == "" || parser.OttPw == "" {
		return fiber.ErrBadRequest
	}

	filter1 := bson.M{"app_id": parser.AppId}
	num, err := getCollection(client, "user").CountDocuments(ctx, filter1)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if num != 1 {
		return fiber.ErrUnauthorized
	}

	filter2 := bson.M{"ott": parser.Ott, "account.id": parser.OttId, "account.pw": parser.OttPw}
	num, err = getCollection(client, "group").CountDocuments(ctx, filter2)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var body struct {
		GroupId string `json:"group_id"`
	}
	switch num {
	case 0:
		account, err := getAccount(parser.Ott, parser.OttId, parser.OttPw)
		if err != nil {
			return err
		}

		var group group
		group.Ott = parser.Ott
		group.Account = *account
		group.UpdateTime = time.Now().Unix()
		group.Members = []member{{
			AppId:   parser.AppId,
			IsAdmin: 1,
		}}

		res, err := getCollection(client, "group").InsertOne(ctx, group)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		body.GroupId = res.InsertedID.(primitive.ObjectID).Hex()
		bodyBytes, err := sonic.Marshal(body)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.Send(bodyBytes)
	case 1:
		filter3 := bson.M{"ott": parser.Ott, "account.id": parser.OttId, "account.pw": parser.OttPw, "members.app_id": parser.AppId}
		num, err := getCollection(client, "group").CountDocuments(ctx, filter3)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if num == 1 {
			return fiber.ErrUnauthorized
		}

		update := bson.M{"$push": bson.M{"members": member{parser.AppId, 0}}, "$set": bson.M{"update_time": time.Now().Unix()}}
		if _, err := getCollection(client, "group").UpdateOne(ctx, filter2, update); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		var bodyBson struct {
			GroupId primitive.ObjectID `bson:"_id"`
		}
		if err = getCollection(client, "group").FindOne(ctx, filter2).Decode(&bodyBson); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		body.GroupId = bodyBson.GroupId.Hex()
		bodyBytes, err := sonic.Marshal(body)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.Send(bodyBytes)
	}

	return fiber.ErrBadRequest
}

func deleteGroup(c *fiber.Ctx) error {
	client, ctx, cancel, err := newClient()
	if err != nil {
		return err
	}
	defer cancel()

	var parser struct {
		AppId string `json:"app_id" bson:"app_id"`
	}
	if err = c.BodyParser(&parser); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if parser.AppId == "" {
		return fiber.ErrBadRequest
	}

	_id, err := primitive.ObjectIDFromHex(c.Params("groupId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var group group
	filter := bson.M{"_id": _id, "members.app_id": parser.AppId}
	if err = getCollection(client, "group").FindOne(ctx, filter).Decode(&group); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if containAdminMembers(group.Members, parser.AppId) {
		if _, err = getCollection(client, "group").DeleteOne(ctx, filter); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(fiber.StatusOK)
	}

	return fiber.ErrNotFound
}

func putGroup(c *fiber.Ctx) error {
	client, ctx, cancel, err := newClient()
	if err != nil {
		return err
	}
	defer cancel()

	var parser struct {
		OttPw      string     `json:"ott_pw" bson:"ott_pw"`
		Payment    payment    `json:"payment" bson:"payment"`
		Membership membership `json:"membership" bson:"membership"`
	}
	if err = c.BodyParser(&parser); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if parser.OttPw == "" || parser.Payment.Type == "" || parser.Payment.Next == 0 || parser.Membership.Type == 0 || parser.Membership.Cost == 0 {
		return fiber.ErrBadRequest
	}

	_id, err := primitive.ObjectIDFromHex(c.Params("groupId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	filter := bson.M{"_id": _id}
	num, err := getCollection(client, "group").CountDocuments(ctx, filter)

	if num == 1 {
		update := bson.M{"$set": bson.M{
			"account.pw":              parser.OttPw,
			"account.payment.type":    parser.Payment.Type,
			"account.payment.detail":  parser.Payment.Detail,
			"account.payment.next":    parser.Payment.Next,
			"account.membership.type": parser.Membership.Type,
			"account.membership.cost": parser.Membership.Cost,
		}}
		if _, err = getCollection(client, "group").UpdateOne(ctx, filter, update); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		return c.SendStatus(fiber.StatusOK)
	}

	return fiber.ErrNotFound
}
