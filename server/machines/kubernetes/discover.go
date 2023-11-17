package kubernetes

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/layer5io/meshery/server/machines"
	"github.com/layer5io/meshery/server/models"
	"github.com/layer5io/meshkit/models/events"
)


type DiscoverAction struct {}

// Execute On Entry and Exit should not return next eventtype i suppose, look again.
func(da *DiscoverAction) ExecuteOnEntry(ctx context.Context, machineCtx interface{}) (machines.EventType, *events.Event, error) {
	user, _ := ctx.Value(models.UserCtxKey).(*models.User)
	sysID, _ := ctx.Value(models.SystemIDKey).(*uuid.UUID)
	userUUID := uuid.FromStringOrNil(user.ID)
	
	eventBuilder := events.NewEvent().ActedUpon(userUUID).WithCategory("connection").WithAction("register").FromSystem(*sysID).FromUser(userUUID) // pass userID and systemID in acted upon first pass user id if we can get context then update with connection Id
	machinectx, err := GetMachineCtx(machineCtx, eventBuilder)
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}
	err = AssignClientSetToContext(machinectx, eventBuilder)
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}
	
	return machines.NoOp, eventBuilder.Build(), nil
}

func(da *DiscoverAction) Execute(ctx context.Context, machineCtx interface{}) (machines.EventType, *events.Event, error) {
	user, _ := ctx.Value(models.UserCtxKey).(*models.User)
	sysID, _ := ctx.Value(models.SystemIDKey).(*uuid.UUID)
	userUUID := uuid.FromStringOrNil(user.ID)

	eventBuilder := events.NewEvent().ActedUpon(userUUID).WithCategory("connection").WithAction("register").FromSystem(*sysID).FromUser(userUUID) // pass userID and systemID in acted upon first pass user id if we can get context then update with connection Id

	machinectx, err := GetMachineCtx(machineCtx, eventBuilder)
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}

	k8sContext := machinectx.K8sContext
	handler := machinectx.clientset

	err = k8sContext.AssignServerID(handler)
	// perofmr event publishinh and err handling
	if err != nil {
		return machines.NoOp, nil, err
	}

	err = k8sContext.AssignVersion(handler)
	// perofmr event publishinh and err handling
	if err != nil {
		return machines.NoOp, nil, err
	}
	token, _ := ctx.Value(models.TokenCtxKey).(string)

	// peform error handling and event publishing
	connection, err := machinectx.Provider.SaveK8sContext(token, machinectx.K8sContext)
	if err != nil {
		return machines.NoOp, eventBuilder.Build(), err
	}
	
	machinectx.log.Debug("exiting execute func from discovered state", connection)

	return machines.Register, nil, nil
}

func(da *DiscoverAction) ExecuteOnExit(ctx context.Context, machineCtx interface{}) (machines.EventType, *events.Event, error) {
	return machines.NoOp, nil, nil
}
