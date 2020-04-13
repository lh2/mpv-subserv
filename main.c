#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "client.h"

extern void Start(char *title);
extern void NewLine(char *line, double sub_start, double sub_end);
extern void Stop();

static int
handle_file_load(mpv_handle *handle)
{
	char *media_title;
	int err;

	err = mpv_get_property(handle, "media-title", MPV_FORMAT_STRING, &media_title);
	if (err != 0) {
		if (!(media_title = malloc(50))) {
			return 1;
		}
		snprintf(media_title, 49, "error getting title %d", err);
	}
	Start(media_title);
	return 0;
}

static void
handle_prop_event(mpv_handle *handle, mpv_event_property *event_prop)
{
	char *value;
	double sub_start = -1, sub_end = -1;
	int err;

	if (strcmp(event_prop->name, "sub-text") != 0 || event_prop->data == NULL) {
		return;
	}

	value = *(char **)(event_prop->data);
	err = mpv_get_property(handle, "sub-start", MPV_FORMAT_DOUBLE, &sub_start);
	if (err != 0) {
		sub_start = -1;
	}
	err = mpv_get_property(handle, "sub-end", MPV_FORMAT_DOUBLE, &sub_end);
	if (err == 0) {
		sub_end = -1;
	}
	NewLine(value, sub_start, sub_end);
}

int
mpv_open_cplugin(mpv_handle *handle)
{
	char *enabled;
	mpv_event *event;
	int err;

	enabled = getenv("MPV_SUBSERV");
	if (enabled == NULL || strcmp(enabled, "1") != 0) {
		printf("mpv-subserv disabled");
		return 0;
	}

	err = mpv_observe_property(handle, 0, "sub-text", MPV_FORMAT_STRING);
	if (err != 0) {
		return err;
	}
	while (1) {
		event = mpv_wait_event(handle, -1);
		if (event->event_id == MPV_EVENT_SHUTDOWN) {
			break;
		}
		switch (event->event_id) {
		case MPV_EVENT_FILE_LOADED:
			if ((err = handle_file_load(handle)) != 0) {
				return err;
			}
			break;
		case MPV_EVENT_PROPERTY_CHANGE:
			handle_prop_event(handle, (mpv_event_property *)event->data);
			break;
		}
	}
	Stop();
	return 0;
}
