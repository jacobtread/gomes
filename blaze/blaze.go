package blaze

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"io"
	"net"
	"strings"
)

var ComponentNames = map[uint16]string{
	0x1:    "Authentication Component",
	0x3:    "Example Component",
	0x4:    "Game Manager Component",
	0x5:    "Redirect Component",
	0x6:    "Play Groups Component",
	0x7:    "Stats Component",
	0x9:    "Util Component",
	0xA:    "Census Data Component",
	0xB:    "Clubs Component",
	0xC:    "Game Report Legacy Component",
	0xD:    "League Component",
	0xE:    "Mail Component",
	0xF:    "Messaging Component",
	0x14:   "Locker Component",
	0x15:   "Rooms Component",
	0x17:   "Tournaments Component",
	0x18:   "Commerce Info Component",
	0x19:   "Association Lists Component",
	0x1B:   "GPS Content Controller Component",
	0x1C:   "Game Reporting Component",
	0x7D0:  "Dynamic Filter Component",
	0x801:  "RSP Component",
	0x7802: "User Sessions Component",
}

var CommandNames = map[uint32]string{
	//Authentication Component
	0x0001000A: "createAccount",
	0x00010014: "updateAccount",
	0x0001001C: "updateParentalEmail",
	0x0001001D: "listUserEntitlements2",
	0x0001001E: "getAccount",
	0x0001001F: "grantEntitlement",
	0x00010020: "listEntitlements",
	0x00010021: "hasEntitlement",
	0x00010022: "getUseCount",
	0x00010023: "decrementUseCount",
	0x00010024: "getAuthToken",
	0x00010025: "getHandoffToken",
	0x00010026: "getPasswordRules",
	0x00010027: "grantEntitlement2",
	0x00010028: "login",
	0x00010029: "acceptTos",
	0x0001002A: "getTosInfo",
	0x0001002B: "modifyEntitlement2",
	0x0001002C: "consumecode",
	0x0001002D: "passwordForgot",
	0x0001002E: "getTermsAndConditionsContent",
	0x0001002F: "getPrivacyPolicyContent",
	0x00010030: "listPersonaEntitlements2",
	0x00010032: "silentLogin",
	0x00010033: "checkAgeReq",
	0x00010034: "getOptIn",
	0x00010035: "enableOptIn",
	0x00010036: "disableOptIn",
	0x0001003C: "expressLogin",
	0x00010046: "logout",
	0x00010050: "createPersona",
	0x0001005A: "getPersona",
	0x00010064: "listPersonas",
	0x0001006E: "loginPersona",
	0x00010078: "logoutPersona",
	0x0001008C: "deletePersona",
	0x0001008D: "disablePersona",
	0x0001008F: "listDeviceAccounts",
	0x00010096: "xboxCreateAccount",
	0x00010098: "originLogin",
	0x000100A0: "xboxAssociateAccount",
	0x000100AA: "xboxLogin",
	0x000100B4: "ps3CreateAccount",
	0x000100BE: "ps3AssociateAccount",
	0x000100C8: "ps3Login",
	0x000100D2: "validateSessionKey",
	0x000100E6: "createWalUserSession",
	0x000100F1: "acceptLegalDocs",
	0x000100F2: "getLegalDocsInfo",
	0x000100F6: "getTermsOfServiceContent",
	0x0001012C: "deviceLoginGuest",
	// Game Manager Component
	0x00040001: "createGame",
	0x00040002: "destroyGame",
	0x00040003: "advanceGameState",
	0x00040004: "setGameSettings",
	0x00040005: "setPlayerCapacity",
	0x00040006: "setPresenceMode",
	0x00040007: "setGameAttributes",
	0x00040008: "setPlayerAttributes",
	0x00040009: "joinGame",
	0x0004000B: "removePlayer",
	0x0004000D: "startMatchmaking",
	0x0004000E: "cancelMatchmaking",
	0x0004000F: "finalizeGameCreation",
	0x00040011: "listGames",
	0x00040012: "setPlayerCustomData",
	0x00040013: "replayGame",
	0x00040014: "returnDedicatedServerToPool",
	0x00040015: "joinGameByGroup",
	0x00040016: "leaveGameByGroup",
	0x00040017: "migrateGame",
	0x00040018: "updateGameHostMigrationStatus",
	0x00040019: "resetDedicatedServer",
	0x0004001A: "updateGameSession",
	0x0004001B: "banPlayer",
	0x0004001D: "updateMeshConnection",
	0x0004001F: "removePlayerFromBannedList",
	0x00040020: "clearBannedList",
	0x00040021: "getBannedList",
	0x00040026: "addQueuedPlayerToGame",
	0x00040027: "updateGameName",
	0x00040028: "ejectHost",
	0x00040050: "*notifyGameUpdated",
	0x00040064: "getGameListSnapshot",
	0x00040065: "getGameListSubscription",
	0x00040066: "destroyGameList",
	0x00040067: "getFullGameData",
	0x00040068: "getMatchmakingConfig",
	0x00040069: "getGameDataFromId",
	0x0004006A: "addAdminPlayer",
	0x0004006B: "removeAdminPlayer",
	0x0004006C: "setPlayerTeam",
	0x0004006D: "changeGameTeamId",
	0x0004006E: "migrateAdminPlayer",
	0x0004006F: "getUserSetGameListSubscription",
	0x00040070: "swapPlayersTeam",
	0x00040096: "registerDynamicDedicatedServerCreator",
	0x00040097: "unregisterDynamicDedicatedServerCreator",
	// Redirector Component
	0x00050001: "getServerInstance",
	// Stats Component
	0x00070001: "getStatDescs",
	0x00070002: "getStats",
	0x00070003: "getStatGroupList",
	0x00070004: "getStatGroup",
	0x00070005: "getStatsByGroup",
	0x00070006: "getDateRange",
	0x00070007: "getEntityCount",
	0x0007000A: "getLeaderboardGroup",
	0x0007000B: "getLeaderboardFolderGroup",
	0x0007000C: "getLeaderboard",
	0x0007000D: "getCenteredLeaderboard",
	0x0007000E: "getFilteredLeaderboard",
	0x0007000F: "getKeyScopesMap",
	0x00070010: "getStatsByGroupAsync",
	0x00070011: "getLeaderboardTreeAsync",
	0x00070012: "getLeaderboardEntityCount",
	0x00070013: "getStatCategoryList",
	0x00070014: "getPeriodIds",
	0x00070015: "getLeaderboardRaw",
	0x00070016: "getCenteredLeaderboardRaw",
	0x00070017: "getFilteredLeaderboardRaw",
	0x00070018: "changeKeyscopeValue",
	// Util Component
	0x00090001: "fetchClientConfig",
	0x00090002: "ping",
	0x00090003: "setClientData",
	0x00090004: "localizeStrings",
	0x00090005: "getTelemetryServer",
	0x00090006: "getTickerServer",
	0x00090007: "preAuth",
	0x00090008: "postAuth",
	0x0009000A: "userSettingsLoad",
	0x0009000B: "userSettingsSave",
	0x0009000C: "userSettingsLoadAll",
	0x0009000E: "deleteUserSettings",
	0x00090014: "filterForProfanity",
	0x00090015: "fetchQosConfig",
	0x00090016: "setClientMetrics",
	0x00090017: "setConnectionState",
	0x00090018: "getPssConfig",
	0x00090019: "getUserOptions",
	0x0009001A: "setUserOptions",
	0x0009001B: "suspendUserPing",
	// Messaging Component
	0x000F0001: "sendMessage",
	0x000F0002: "fetchMessages",
	0x000F0003: "purgeMessages",
	0x000F0004: "touchMessages",
	0x000F0005: "getMessages",
	// Association Lists Component
	0x00190001: "addUsersToList",
	0x00190002: "removeUsersFromList",
	0x00190003: "clearLists",
	0x00190004: "setUsersToList",
	0x00190005: "getListForUser",
	0x00190006: "getLists",
	0x00190007: "subscribeToLists",
	0x00190008: "unsubscribeFromLists",
	0x00190009: "getConfigListsInfo",
	// Game Reporting Component
	0x001C0001: "submitGameReport",
	0x001C0002: "submitOfflineGameReport",
	0x001C0003: "submitGameEvents",
	0x001C0004: "getGameReportQuery",
	0x001C0005: "getGameReportQueriesList",
	0x001C0006: "getGameReports",
	0x001C0007: "getGameReportView",
	0x001C0008: "getGameReportViewInfo",
	0x001C0009: "getGameReportViewInfoList",
	0x001C000A: "getGameReportTypes",
	0x001C000B: "updateMetric",
	0x001C000C: "getGameReportColumnInfo",
	0x001C000D: "getGameReportColumnValues",
	0x001C0064: "submitTrustedMidGameReport",
	0x001C0065: "submitTrustedEndGameReport",
	// User Sessions Component
	0x78020003: "fetchExtendedData",
	0x78020005: "updateExtendedDataAttribute",
	0x78020008: "updateHardwareFlags",
	0x7802000C: "lookupUser",
	0x7802000D: "lookupUsers",
	0x7802000E: "lookupUsersByPrefix",
	0x78020014: "updateNetworkInfo",
	0x78020017: "lookupUserGeoIPData",
	0x78020018: "overrideUserGeoIPData",
	0x78020019: "updateUserSessionClientData",
	0x7802001A: "setUserInfoAttribute",
	0x7802001B: "resetUserGeoIPData",
	0x78020020: "lookupUserSessionId",
	0x78020021: "fetchLastLocaleUsedAndAuthError",
	0x78020022: "fetchUserFirstLastAuthTime",
	0x78020023: "resumeSession",
}

type Connection struct {
	PacketBuff
	net.Conn
}

type PacketBuff struct {
	*bytes.Buffer
}

type Packet struct {
	Length    uint16
	Component uint16
	Command   uint16
	Error     uint16
	QType     uint16
	Id        uint16
	ExtLength uint16
	Content   []byte
}

// UInt16 reads an uint16 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) UInt16() uint16 {
	var out uint16
	_ = binary.Read(b, binary.BigEndian, &out)
	return out
}

// UInt32 reads an uint32 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) UInt32() uint32 {
	var out uint32
	_ = binary.Read(b, binary.BigEndian, &out)
	return out
}

// Float64 reads a float64 from the provided packet buffer using the
// big endian byte order
func (b *PacketBuff) Float64() float64 {
	var out float64
	_ = binary.Read(b, binary.BigEndian, &out)
	return out
}

// WriteVarInt writes a var int to the packet buffer
func (b *PacketBuff) WriteVarInt(value int64) {
	ux := uint64(value) << 1
	if value < 0 {
		ux = ^ux
	}
	i := 0
	for ux >= 0x80 {
		_ = b.WriteByte(byte(ux) | 0x80)
		ux >>= 7
		i++
	}
	_ = b.WriteByte(byte(ux))
}

// ReadVarInt reads a var int from the packet buffer
func (b *PacketBuff) ReadVarInt() uint64 {
	var x uint64
	var s uint
	for i := 0; i < 10; i++ {
		b, err := b.ReadByte()
		if err != nil {
			return x
		}
		if b < 0x80 {
			if i == 9 && b > 1 {
				return x
			}
			return x | uint64(b)<<s
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return x
}

// WriteNum takes any number type and writes it to the packet
func (b *PacketBuff) WriteNum(value any) {
	_ = binary.Write(b, binary.BigEndian, value)
}

// ReadString reads a string from the buffer
func (b *PacketBuff) ReadString() string {
	l := b.ReadVarInt()
	buf := make([]byte, l)
	_, _ = io.ReadFull(b, buf)
	_, _ = b.ReadByte() // Strings end with a zero byte
	return string(buf)
}

// WriteString writes a string to the buffer
func (b *PacketBuff) WriteString(value string) {
	var l int
	if strings.HasSuffix(value, "\x00") {
		l = len(value)
	} else {
		l = len(value) + 1
	}
	b.WriteVarInt(int64(l))
	_, _ = b.Write([]byte(value))
	_ = b.WriteByte(0)
}

// ReadPacket reads a game packet from the provided packet reader
func (b *PacketBuff) ReadPacket() *Packet {
	packet := Packet{
		Length:    b.UInt16(),
		Component: b.UInt16(),
		Command:   b.UInt16(),
		Error:     b.UInt16(),
		QType:     b.UInt16(),
		Id:        b.UInt16(),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	by := make([]byte, l)        // Create an empty byte array for the content
	_, err := io.ReadFull(b, by) // Read all the content bytes
	if err != nil {
		return nil
	}
	packet.Content = by
	return &packet
}

// ReadPacketHeading reads a game packet from the provided packet reader.
// but only reads the heading portion of the packet skips over the packet
// contents.
func (b *PacketBuff) ReadPacketHeading() *Packet {
	packet := Packet{
		Length:    b.UInt16(),
		Component: b.UInt16(),
		Command:   b.UInt16(),
		Error:     b.UInt16(),
		QType:     b.UInt16(),
		Id:        b.UInt16(),
	}
	if (packet.QType * 0x10) != 0 {
		packet.ExtLength = b.UInt16()
	} else {
		packet.ExtLength = 0
	}
	// Calculate the total size with the extension length
	l := int32(packet.Length) + (int32(packet.ExtLength) << 16)
	by := make([]byte, l) // Create an empty byte array in place of the content
	packet.Content = by
	return &packet
}

func (b *PacketBuff) ReadAllPackets() *list.List {
	out := list.New()
	for b.Len() > 0 {
		out.PushBack(b.ReadPacket())
	}
	return out
}

func (b *PacketBuff) EncodePacket(comp uint16, cmd uint16, err uint16, qType uint16, id uint16, content list.List) []byte {
	buf := &PacketBuff{Buffer: &bytes.Buffer{}}
	contentBuff := &PacketBuff{Buffer: &bytes.Buffer{}}
	for l := content.Front(); l != nil; l = l.Next() {
		WriteTdf(contentBuff, l.Value.(Tdf))
	}
	c := contentBuff.Bytes()
	l := len(c)

	_ = buf.WriteByte(byte((l & 0xFFFF) >> 8))
	_ = buf.WriteByte(byte(l & 0xFF))
	_ = buf.WriteByte(0)
	_ = binary.Write(buf, binary.BigEndian, comp)
	_ = binary.Write(buf, binary.BigEndian, cmd)
	_ = binary.Write(buf, binary.BigEndian, err)

	buf.WriteByte(byte(qType >> 8))
	if l > 0xFFFF {
		buf.WriteByte(0x10)
	} else {
		buf.WriteByte(0x00)
	}

	_ = binary.Write(buf, binary.BigEndian, id)

	if l > 0xFFFF {
		buf.WriteByte(byte((l & 0xFF000000) >> 24))
		buf.WriteByte(byte((l & 0x00FF0000) >> 16))
	}

	_, _ = buf.Write(c)
	return buf.Bytes()
}

func (b *PacketBuff) EncodePacketRaw(packet Packet) []byte {
	buf := &PacketBuff{Buffer: &bytes.Buffer{}}
	_ = buf.WriteByte(byte(packet.Length >> 8))
	_ = buf.WriteByte(byte(packet.Length & 0xFF))
	_ = buf.WriteByte(0)
	_ = binary.Write(buf, binary.BigEndian, packet.Component)
	_ = binary.Write(buf, binary.BigEndian, packet.Command)
	_ = binary.Write(buf, binary.BigEndian, packet.Error)
	_ = binary.Write(buf, binary.BigEndian, packet.QType)
	_ = binary.Write(buf, binary.BigEndian, packet.Id)
	if (packet.QType & 0x10) != 0 {
		buf.WriteByte(byte(packet.ExtLength >> 8))
		buf.WriteByte(byte(packet.ExtLength & 0xFF))
	}
	_, _ = buf.Write(packet.Content)
	return buf.Bytes()
}

func (packet *Packet) ReadContent() *list.List {
	buff := PacketBuff{Buffer: bytes.NewBuffer(packet.Content)}
	out := list.New()
	for buff.Len() > 0 {
		out.PushBack(buff.ReadTdf())
	}
	return out
}
