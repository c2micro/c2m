package pools

import (
	"context"

	"github.com/c2micro/c2mshr/defaults"
	operatorv1 "github.com/c2micro/c2mshr/proto/gen/operator/v1"
	"github.com/c2micro/c2m/internal/ent"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// конвертим модель Listener (БД) в ListenerResponse (proto)
func ToListenerResponse(listener *ent.Listener) *operatorv1.ListenerResponse {
	return &operatorv1.ListenerResponse{
		Lid:   int64(listener.ID),
		Name:  wrapperspb.String(listener.Name),
		Ip:    wrapperspb.String(listener.IP.String()),
		Port:  wrapperspb.UInt32(uint32(listener.Port)),
		Note:  wrapperspb.String(listener.Note),
		Last:  timestamppb.New(listener.Last),
		Color: wrapperspb.UInt32(listener.Color),
	}
}

// конвертим массив моделей Listener (БД) в ListenersResponse (proto)
func ToListenersResponse(listeners []*ent.Listener) *operatorv1.ListenersResponse {
	temp := make([]*operatorv1.ListenerResponse, 0)
	for _, listener := range listeners {
		temp = append(temp, ToListenerResponse(listener))
	}
	return &operatorv1.ListenersResponse{
		Listeners: temp,
	}
}

// конвертим модель Listener (БД) в ListenerNoteResponse (proto)
func ToListenerNoteResponse(listener *ent.Listener) *operatorv1.ListenerNoteResponse {
	return &operatorv1.ListenerNoteResponse{
		Lid:  int64(listener.ID),
		Note: wrapperspb.String(listener.Note),
	}
}

// конвертим модель Credential (БД) в CredntialNoteResponse (proto)
func ToCredentialNoteResponse(credential *ent.Credential) *operatorv1.CredentialNoteResponse {
	return &operatorv1.CredentialNoteResponse{
		Cid: int64(credential.ID),
		Note: &wrapperspb.StringValue{
			Value: credential.Note,
		},
	}
}

// конвертим модель Listener (БД) в ListenerInfoResponse (proto)
func ToListenerInfoResponse(listener *ent.Listener) *operatorv1.ListenerInfoResponse {
	return &operatorv1.ListenerInfoResponse{
		Lid:  int64(listener.ID),
		Name: wrapperspb.String(listener.Name),
		Ip:   wrapperspb.String(listener.IP.String()),
		Port: wrapperspb.UInt32(uint32(listener.Port)),
	}
}

// конвертим модель Beacon (БД) в BeaconResponse (proto)
func ToBeaconResponse(beacon *ent.Beacon) (*operatorv1.BeaconResponse, error) {
	listener, err := beacon.Edges.ListenerOrErr()
	if err != nil {
		return nil, err
	}
	return &operatorv1.BeaconResponse{
		Bid:        beacon.Bid,
		Lid:        int64(listener.ID),
		ExtIp:      wrapperspb.String(beacon.ExtIP.String()),
		IntIp:      wrapperspb.String(beacon.IntIP.String()),
		Os:         uint32(beacon.Os),
		OsMeta:     wrapperspb.String(beacon.OsMeta),
		Hostname:   wrapperspb.String(beacon.Hostname),
		Username:   wrapperspb.String(beacon.Username),
		Domain:     wrapperspb.String(beacon.Domain),
		Privileged: wrapperspb.Bool(beacon.Privileged),
		ProcName:   wrapperspb.String(beacon.ProcessName),
		Pid:        wrapperspb.UInt32(beacon.Pid),
		Arch:       uint32(beacon.Arch),
		Sleep:      beacon.Sleep,
		Jitter:     uint32(beacon.Jitter),
		Caps:       beacon.Caps,
		Color:      wrapperspb.UInt32(beacon.Color),
		Note:       wrapperspb.String(beacon.Note),
		First:      timestamppb.New(beacon.First),
		Last:       timestamppb.New(beacon.Last),
	}, nil
}

// конвертим массив моделей Beacon (БД) в BeaconsResponse (proto)
func ToBeaconsResponse(beacons []*ent.Beacon) *operatorv1.BeaconsResponse {
	temp := make([]*operatorv1.BeaconResponse, 0)
	for _, beacon := range beacons {
		beaconResponse, err := ToBeaconResponse(beacon)
		if err != nil {
			// пропускаем бикон
			continue
		}
		temp = append(temp, beaconResponse)
	}
	return &operatorv1.BeaconsResponse{
		Beacons: temp,
	}
}

// конвертим модель Beacon (БД) в BeaconLastResponse (proto)
func ToBeaconLastResponse(beacon *ent.Beacon) *operatorv1.BeaconLastResponse {
	return &operatorv1.BeaconLastResponse{
		Bid:  beacon.Bid,
		Last: timestamppb.New(beacon.Last),
	}
}

// конвертим модель Operator (БД) в OperatorResponse (proto)
func ToOperatorResponse(operator *ent.Operator) *operatorv1.OperatorResponse {
	return &operatorv1.OperatorResponse{
		Username: operator.Username,
		Color:    wrapperspb.UInt32(operator.Color),
		Last:     timestamppb.New(operator.Last),
	}
}

// конвертим массив моделей Operator (БД) в OperatorsResponse (proto)
func ToOperatorsResponse(operators []*ent.Operator) *operatorv1.OperatorsResponse {
	temp := make([]*operatorv1.OperatorResponse, 0)
	for _, operator := range operators {
		temp = append(temp, &operatorv1.OperatorResponse{
			Username: operator.Username,
			Color:    wrapperspb.UInt32(operator.Color),
			Last:     timestamppb.New(operator.Last),
		})
	}
	return &operatorv1.OperatorsResponse{
		Operators: temp,
	}
}

// конвертим модель Chat (БД) в ChatResponse (proto)
func ToChatResponse(chatMessage *ent.Chat) (*operatorv1.ChatResponse, error) {
	author := defaults.ChatSrvFrom
	if o, err := chatMessage.Edges.OperatorOrErr(); err != nil {
		if !ent.IsNotFound(err) {
			if ent.IsNotLoaded(err) {
				o, err = chatMessage.QueryOperator().Only(context.Background())
				if err != nil {
					return nil, err
				} else {
					author = o.Username
				}
			} else {
				return nil, err
			}
		}
	} else {
		author = o.Username
	}
	return &operatorv1.ChatResponse{
		CreatedAt: timestamppb.New(chatMessage.CreatedAt),
		From:      author,
		Message:   chatMessage.Message,
	}, nil
}

// конвертим массив моделей Chat (БД) в ChatMessagesResponse (proto)
func ToChatMessagesResponse(chatMessages []*ent.Chat) *operatorv1.ChatMessagesResponse {
	temp := make([]*operatorv1.ChatResponse, 0)

	for _, chatMessage := range chatMessages {
		chatResponse, err := ToChatResponse(chatMessage)
		if err != nil {
			// пропускаем сообщение
			continue
		}
		temp = append(temp, chatResponse)
	}

	return &operatorv1.ChatMessagesResponse{
		Messages: temp,
	}
}

// конвертим модель Credential (БД) в CredentialResponse (proto)
func ToCredentialResponse(credential *ent.Credential) *operatorv1.CredentialResponse {
	return &operatorv1.CredentialResponse{
		Cid:       int64(credential.ID),
		Username:  wrapperspb.String(credential.Username),
		Password:  wrapperspb.String(credential.Secret),
		Realm:     wrapperspb.String(credential.Realm),
		Host:      wrapperspb.String(credential.Host),
		CreatedAt: timestamppb.New(credential.CreatedAt),
		Note:      wrapperspb.String(credential.Note),
		Color:     wrapperspb.UInt32(credential.Color),
	}
}

// конвертим массив моделей Credential (БД) в CredentialsResponse (proto)
func ToCredentialsResponse(credentials []*ent.Credential) *operatorv1.CredentialsResponse {
	temp := make([]*operatorv1.CredentialResponse, 0)
	for _, credential := range credentials {
		temp = append(temp, ToCredentialResponse(credential))
	}
	return &operatorv1.CredentialsResponse{
		Credentials: temp,
	}
}
