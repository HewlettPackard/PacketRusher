package WorkLoadModel

/*
#cgo CFLAGS: -I /home/lucas-baleeiro/gsl/include
#cgo LDFLAGS: -L /home/lucas-baleeiro/gsl/lib -lgsl -lgslcblas -lm
#include<gsl/gsl_rng.h>
#include<gsl/gsl_randist.h>

unsigned int randomNumberFromPoisson( double mu, long seed) {
    const gsl_rng_type* T;
    gsl_rng* r;
    unsigned int dist;

	gsl_rng_env_setup();
	T = gsl_rng_default;
	r = gsl_rng_alloc(T);

    gsl_rng_set(r,seed);
    dist = gsl_ran_poisson (r, mu);

	gsl_rng_free (r);

	return dist;
}
*/
import "C"

// calculate some random numbers from the Poisson distribution
func poissonDistribution(mean float64, length int) (distPoisson []uint) {
	distPoisson = make([]uint, length)

	for i := 1; i <= length; i++ {
		aux := C.randomIntengerFromPoisson(C.double(mean), C.long(i))
		distPoisson[i-1] = uint(aux)
	}
	return
}
