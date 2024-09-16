#ifndef PAMAC_HELPERS
#define PAMAC_HELPERS

#include <pamac.h>
#include <stdio.h>

typedef struct {
    PamacDatabase* db;
} kernel_manager;

struct kernels {
    PamacPackage** data;
    int len;
};

static inline kernel_manager* create_manager() {
    return (kernel_manager*)calloc(1, sizeof(kernel_manager));
}

static inline void create_db(kernel_manager* mgr) {
    PamacConfig* cfg = pamac_config_new("/etc/pamac.conf");
    mgr->db = pamac_database_new(cfg);
    pamac_database_refresh(mgr->db);
}

static inline struct kernels get_kernels(PamacDatabase* db) {
    GPtrArray* pkgs =
        pamac_database_search_repos_pkgs(db, "^linux([0-9][0-9]?([0-9])|[0-9][0-9]?([0-9])-rt)$");

    struct kernels kernels;
    kernels.data = calloc(pkgs->len, sizeof(PamacPackage*));
    kernels.len = pkgs->len;

    for (size_t cnt = 0; cnt < kernels.len; cnt++) {
        kernels.data[cnt] = (PamacPackage*)pkgs->pdata[cnt];
    }

    free(pkgs);
    return kernels;
}

static inline void install_callback(GObject* source_object, GAsyncResult* res, gpointer data) {
    PamacTransaction* transaction = (PamacTransaction*)source_object;
    gboolean success = pamac_transaction_run_finish(transaction, res);

    printf("XXX INSTALL CALLBACK!\n");
    fflush(stdout);

    op_callback(success);
}

static inline void install_kernel(PamacDatabase* db, char const* name) {
    PamacTransaction* transaction = pamac_transaction_new(db);

    char headers[1024];
    strcpy(headers, name);
    strcat(headers, "-headers");

    pamac_transaction_add_pkg_to_install(transaction, name);
    pamac_transaction_add_pkg_to_install(transaction, headers);

    pamac_transaction_run_async(transaction, install_callback, NULL);
    printf("XXX END INSTALL KERNEL\n");
    fflush(stdout);
}

static inline void remove_callback(GObject* source_object, GAsyncResult* res, gpointer data) {
    PamacTransaction* transaction = (PamacTransaction*)source_object;
    gboolean success = pamac_transaction_run_finish(transaction, res);
    printf("XXX REMOVE CALLBACK!\n");
    fflush(stdout);

    op_callback(success);
}

static inline void remove_kernel(PamacDatabase* db, char const* name) {
    PamacTransaction* transaction = pamac_transaction_new(db);

    char headers[1024];
    strcpy(headers, name);
    strcat(headers, "-headers");

    pamac_transaction_add_pkg_to_remove(transaction, name);
    pamac_transaction_add_pkg_to_remove(transaction, headers);

    pamac_transaction_run_async(transaction, remove_callback, NULL);
    printf("XXX END REMOVE KERNEL\n");
    fflush(stdout);
}

static inline void free_kernels(struct kernels* kernels) {
    free(kernels->data);
}

#endif
