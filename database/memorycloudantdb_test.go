package database_test

import (
	. "github.com/amalgam8/controller/database"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type record struct {
	ID      string `json:"_id"`
	Rev     string `json:"_rev"`
	Counter int    `json:"counter"`
}

type allRecords struct {
	Docs []struct {
		Fields record `json:"doc"`
	} `json:"rows"`
	TotalRows int `json:"total_rows"`
}

func (ar *allRecords) GetEntries() []Entry {
	var entries []Entry
	for _, entry := range ar.Docs {
		entries = append(entries, &entry.Fields)
	}
	return entries
}

func (r *record) IDRev() (string, string) {
	return r.ID, r.Rev
}

func (r *record) SetRev() {}

func (r *record) SetIV(iv string) {}

func (r *record) GetIV() string {
	return "mock_IV"
}

var _ = Describe("Memorycloudantdb", func() {

	var (
		db   CloudantDB
		test record
	)

	Describe("Interating with the in-memory database ", func() {

		Context("creating a new instance", func() {
			It("should return a CloudantDB instance", func() {
				db = NewMemoryCloudantDB()
				Expect(db).NotTo(BeNil())
			})
		})

		Context("adding a record", func() {
			It("should not error", func() {
				db = NewMemoryCloudantDB()
				test = record{
					ID:  "001",
					Rev: "111",
				}
				err := db.InsertEntry(&test)
				Expect(err).ShouldNot(HaveOccurred())

				By("reading all ids should have 1 record")
				ids, _ := db.ReadKeys()
				Expect(ids).Should(HaveLen(1))

			})
		})

		Context("updating a record", func() {

			JustBeforeEach(func() {
				db = NewMemoryCloudantDB()
				test = record{
					ID:      "001",
					Rev:     "111",
					Counter: 1,
				}
				db.InsertEntry(&test)
				ids, _ := db.ReadKeys()
				Expect(ids).Should(HaveLen(1))
			})

			It("should have 1 record", func() {
				update := record{
					ID:      "001",
					Rev:     "222",
					Counter: 2,
				}

				db.InsertEntry(&update)

				By("reading the record")
				response := new(record)
				err := db.ReadEntry("001", response)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.Counter).To(Equal(2))
			})
		})

		Context("reading a record", func() {

			JustBeforeEach(func() {
				db = NewMemoryCloudantDB()
				test = record{
					ID:      "001",
					Rev:     "111",
					Counter: 3,
				}
				err := db.InsertEntry(&test)
				Expect(err).ShouldNot(HaveOccurred())

			})

			It("should Error 'Key not found'", func() {
				response := new(record)
				err := db.ReadEntry("000", response)
				Expect(err).Should(HaveOccurred())
				// check for 404
				if de, ok := err.(*DBError); ok {
					Expect(de.StatusCode).To(Equal(404))
				} else {
					// should not have anything other than a cloudant eror
					Expect(err).ShouldNot(HaveOccurred())
				}
			})

			It("should return the record", func() {
				response := new(record)
				err := db.ReadEntry("001", response)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(response.Counter).To(Equal(3))
			})

		})

		Context("removing a record", func() {

			JustBeforeEach(func() {
				db = NewMemoryCloudantDB()
				test = record{
					ID:  "001",
					Rev: "111",
				}
				db.InsertEntry(&test)
				ids, _ := db.ReadKeys()
				Expect(ids).Should(HaveLen(1))
			})

			It("should Error, 'Key not found'", func() {
				err := db.DeleteEntry("000")
				Expect(err).Should(HaveOccurred())
				// check for 404
				if de, ok := err.(*DBError); ok {
					Expect(de.StatusCode).To(Equal(404))
				} else {
					// should not have anything other than a cloudant eror
					Expect(err).ShouldNot(HaveOccurred())
				}
			})
			It("should have 0 records", func() {
				db.DeleteEntry("001")
				ids, err := db.ReadKeys()
				Expect(ids).Should(HaveLen(0))
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("should have 0 records", func() {
				db.DeleteEntry("001")
				ids, _ := db.ReadKeys()
				Expect(ids).Should(HaveLen(0))
			})
		})

		Context("read all records with content", func() {

			BeforeEach(func() {
				db = NewMemoryCloudantDB()
				test1 := record{
					ID:      "001",
					Rev:     "111",
					Counter: 1,
				}
				test2 := record{
					ID:      "002",
					Rev:     "222",
					Counter: 2,
				}

				test3 := record{
					ID:      "003",
					Rev:     "333",
					Counter: 3,
				}

				db.InsertEntry(&test1)
				db.InsertEntry(&test2)
				db.InsertEntry(&test3)
				ids, _ := db.ReadKeys()
				Expect(ids).Should(HaveLen(3))
			})

			It("Print", func() {
				allDocs := new(allRecords)
				err := db.ReadAllDocsContent(allDocs)
				Expect(err).ShouldNot(HaveOccurred())

				By("TotalRows should be 3")
				Expect(allDocs.TotalRows).To(Equal(3))
			})
		})

		Context("check if database exists", func() {
			JustBeforeEach(func() {
				db = NewMemoryCloudantDB()
			})

			It("should exist", func() {
				exists, err := db.DBExists("somedb")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(exists).To(Equal(true))
			})
		})

	})
})
