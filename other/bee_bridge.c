//go:build ignore
// +build ignore

#include <windows.h>
#include "go_plugin.h"

char *__stdcall bee_init(char *a) { return GoBeeInit(a); }
void __stdcall bee_enable(char *a) { GoBeeEnable(a); }
void __stdcall bee_disable(char *a) { GoBeeDisable(a); }
void __stdcall bee_unload(char *a) { GoBeeUnload(a); }
void __stdcall bee_settings(char *a) { GoBeeSettings(a); }
int __stdcall bee_channel_dm(char*a,char*b,char*c,char*d,char*e,char*f){return GoBeeChannelDM(a,b,c,d,e,f);}
int __stdcall bee_channel_msg(char*a,char*b,char*c,char*d,char*e,char*f){return GoBeeChannelMessage(a,b,c,d,e,f);}
int __stdcall bee_channel_event(char*a,char*b,char*c,char*d,char*e,char*f,char*g){return GoBeeChannelEvent(a,b,c,d,e,f,g);}
int __stdcall bee_private_msg(char*a,char*b,char*c,char*d){return GoBeePrivateMessage(a,b,c,d);}
int __stdcall bee_group_msg(char*a,char*b,char*c,char*d,char*e){return GoBeeGroupMessage(a,b,c,d,e);}
int __stdcall bee_common_event(char*a,char*b,char*c,char*d,char*e,char*f){return GoBeeCommonEvent(a,b,c,d,e,f);}

