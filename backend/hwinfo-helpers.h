#ifndef LIBHD_HELPERS
#define LIBHD_HELPERS

#include <assert.h>
#include <hd.h>
#include <malloc.h>
#include <stdio.h>

typedef struct {
} hw_manager;

typedef enum {
    HWUSB = 0,
    HWPCI,
} hwtype;

typedef struct {
    hd_data_t* data;
    hd_t* first;
} hw_list;

static inline hw_list hw_get_devices(hwtype type) {
    hw_list list;
    hd_hw_item_t hw;

    if (type == HWUSB) {
        hw = hw_usb;
    } else {
        assert(type == HWPCI);
        hw = hw_pci;
    }

    list.data = (hd_data_t*)calloc(1, sizeof *list.data);
    list.first = hd_list(list.data, hw, 1, NULL);

    printf("XXX hw_get_devices END\n");
    fflush(stdout);

    return list;
}

#endif
