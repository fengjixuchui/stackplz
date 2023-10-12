typedef unsigned long uint64_t;

struct UnwindBuf {
    uint64_t abi;
    uint64_t regs[33];
    uint64_t size;
    uint64_t dyn_size;
};

const char* get_stack(char* dl_path, char* map_buffer, uint64_t reg_mask, void* unwind_buf, void* stack_buf);